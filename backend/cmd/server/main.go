package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // lets the TZ environment variable (e.g. Asia/Kolkata) work on any host

	"employee-management/internal/config"
	"employee-management/internal/handler"
	"employee-management/internal/repository"
	"employee-management/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/joho/godotenv"
)

func main() {
	// .env is optional; real environment variables take precedence.
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := waitForDB(db, 10, 2*time.Second); err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	log.Println("connected to PostgreSQL")

	repo := repository.NewEmployeeRepository(db)
	svc := service.NewEmployeeService(repo)
	h := handler.NewEmployeeHandler(svc)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.CORSOrigin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
		MaxAge:       12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	h.RegisterRoutes(router.Group("/api/v1"))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
}

// waitForDB retries Ping so the API can start while PostgreSQL is still booting.
func waitForDB(db *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 1; i <= attempts; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		log.Printf("database not ready (attempt %d/%d): %v", i, attempts, err)
		time.Sleep(delay)
	}
	return err
}
