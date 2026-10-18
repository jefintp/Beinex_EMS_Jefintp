import { Injectable, signal } from '@angular/core';

export interface Toast {
  type: 'success' | 'error';
  text: string;
}

/** Tiny toast store; rendered once by AppComponent. */
@Injectable({ providedIn: 'root' })
export class NotificationService {
  readonly message = signal<Toast | null>(null);
  private timer: ReturnType<typeof setTimeout> | undefined;

  success(text: string): void {
    this.show({ type: 'success', text });
  }

  error(text: string): void {
    this.show({ type: 'error', text });
  }

  dismiss(): void {
    clearTimeout(this.timer);
    this.message.set(null);
  }

  private show(toast: Toast): void {
    clearTimeout(this.timer);
    this.message.set(toast);
    this.timer = setTimeout(() => this.message.set(null), 4000);
  }
}
