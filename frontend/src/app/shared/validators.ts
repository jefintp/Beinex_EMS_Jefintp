import { AbstractControl, ValidationErrors } from '@angular/forms';

// Angular's built-in Validators.email accepts "a@b"; require a dotted domain as well.
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

export function emailFormat(control: AbstractControl): ValidationErrors | null {
  const value = String(control.value ?? '').trim();
  return !value || EMAIL_RE.test(value) ? null : { emailFormat: true };
}

/** Today's date in the browser's local time, as YYYY-MM-DD (the format of <input type="date">). */
export function localToday(): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${now.getFullYear()}-${month}-${day}`;
}

/** Rejects dates after today. Both sides are YYYY-MM-DD, so string comparison works. */
export function notFutureDate(control: AbstractControl): ValidationErrors | null {
  const value = String(control.value ?? '');
  return value && value > localToday() ? { futureDate: true } : null;
}

/** Rejects values that are only whitespace (empty values are left to `required`). */
export function notBlank(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  return typeof value === 'string' && value.length > 0 && value.trim().length === 0
    ? { blank: true }
    : null;
}
