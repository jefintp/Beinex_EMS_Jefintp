import { DatePipe } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { BehaviorSubject, catchError, combineLatest, debounceTime, of, startWith, switchMap, tap } from 'rxjs';

import { parseApiError } from '../../core/api-error';
import { NotificationService } from '../../core/notification.service';
import { ConfirmDialogComponent } from '../../shared/confirm-dialog/confirm-dialog.component';
import { DEPARTMENTS, Employee, STATUSES } from '../employee.model';
import { EmployeeService } from '../employee.service';

@Component({
  selector: 'app-employee-list',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink, DatePipe, ConfirmDialogComponent],
  templateUrl: './employee-list.component.html',
})
export class EmployeeListComponent {
  private readonly fb = inject(FormBuilder);
  private readonly api = inject(EmployeeService);
  private readonly notify = inject(NotificationService);

  readonly departments = DEPARTMENTS;
  readonly statuses = STATUSES;

  readonly filters = this.fb.nonNullable.group({
    search: '',
    department: '',
    status: '',
  });

  readonly employees = signal<Employee[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly pendingDelete = signal<Employee | null>(null);
  readonly deleting = signal(false);
  readonly deleteError = signal<string | null>(null);

  /** Emits to re-run the current query (retry, or after a delete). */
  private readonly refresh$ = new BehaviorSubject<void>(undefined);

  constructor() {
    combineLatest([
      // Debounce typing; the initial value is emitted immediately by startWith.
      this.filters.valueChanges.pipe(debounceTime(300), startWith(this.filters.getRawValue())),
      this.refresh$,
    ])
      .pipe(
        tap(() => {
          this.loading.set(true);
          this.error.set(null);
        }),
        // switchMap cancels an in-flight request when the filters change again.
        switchMap(([filter]) =>
          this.api.list(filter).pipe(
            catchError((err) => {
              this.error.set(parseApiError(err).message);
              return of<Employee[]>([]);
            }),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe((employees) => {
        this.employees.set(employees);
        this.loading.set(false);
      });
  }

  hasActiveFilters(): boolean {
    const { search, department, status } = this.filters.getRawValue();
    return search.trim() !== '' || department !== '' || status !== '';
  }

  clearFilters(): void {
    this.filters.reset();
  }

  reload(): void {
    this.refresh$.next();
  }

  askDelete(employee: Employee): void {
    this.deleteError.set(null);
    this.pendingDelete.set(employee);
  }

  cancelDelete(): void {
    if (this.deleting()) return;
    this.pendingDelete.set(null);
  }

  confirmDelete(): void {
    const target = this.pendingDelete();
    if (!target || this.deleting()) return;

    this.deleting.set(true);
    this.deleteError.set(null);

    this.api.delete(target.id).subscribe({
      next: () => {
        this.deleting.set(false);
        this.pendingDelete.set(null);
        this.notify.success(`${target.name} was deleted.`);
        this.reload();
      },
      error: (err) => {
        this.deleting.set(false);
        this.deleteError.set(parseApiError(err).message);
      },
    });
  }
}
