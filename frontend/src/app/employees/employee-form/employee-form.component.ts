import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { parseApiError } from '../../core/api-error';
import { NotificationService } from '../../core/notification.service';
import { emailFormat, localToday, notBlank, notFutureDate } from '../../shared/validators';
import { DEPARTMENTS, EmployeePayload, EmployeeStatus, STATUSES } from '../employee.model';
import { EmployeeService } from '../employee.service';

type FieldName = 'name' | 'email' | 'department' | 'designation' | 'joining_date' | 'status';

const FIELD_LABELS: Record<FieldName, string> = {
  name: 'Name',
  email: 'Email',
  department: 'Department',
  designation: 'Designation',
  joining_date: 'Joining date',
  status: 'Status',
};

/** Handles both "Add employee" (`/employees/new`) and "Edit employee" (`/employees/:id/edit`). */
@Component({
  selector: 'app-employee-form',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './employee-form.component.html',
})
export class EmployeeFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly api = inject(EmployeeService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly notify = inject(NotificationService);

  readonly departments = DEPARTMENTS;
  readonly statuses = STATUSES;
  /** Latest selectable joining date (the date picker greys out anything later). */
  readonly today = localToday();

  readonly employeeId = signal<number | null>(null);
  readonly isEdit = computed(() => this.employeeId() !== null);

  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly loadError = signal<string | null>(null);
  readonly submitError = signal<string | null>(null);

  readonly form = this.fb.nonNullable.group({
    name: ['', [Validators.required, notBlank, Validators.minLength(2), Validators.maxLength(100)]],
    email: ['', [Validators.required, emailFormat, Validators.maxLength(254)]],
    department: ['', [Validators.required]],
    designation: ['', [Validators.required, notBlank, Validators.maxLength(100)]],
    joining_date: ['', [Validators.required, notFutureDate]],
    status: ['active' as EmployeeStatus, [Validators.required]],
  });

  ngOnInit(): void {
    const idParam = this.route.snapshot.paramMap.get('id');
    if (idParam === null) return; // add mode

    const id = Number(idParam);
    if (!Number.isInteger(id) || id <= 0) {
      this.loadError.set('Invalid employee id.');
      return;
    }

    this.employeeId.set(id);
    this.loading.set(true);
    this.api.get(id).subscribe({
      next: (employee) => {
        this.form.patchValue({
          name: employee.name,
          email: employee.email,
          department: employee.department,
          designation: employee.designation,
          // <input type="date"> needs YYYY-MM-DD, even if the API sends a full timestamp.
          joining_date: employee.joining_date.slice(0, 10),
          status: employee.status,
        });
        this.loading.set(false);
      },
      error: (err) => {
        this.loadError.set(parseApiError(err).message);
        this.loading.set(false);
      },
    });
  }

  /** Message to show under a field, or null while it is valid / untouched. */
  errorFor(field: FieldName): string | null {
    const control = this.form.controls[field];
    const errors = control.errors;
    if (!errors || !(control.touched || control.dirty)) return null;

    const label = FIELD_LABELS[field];
    if (errors['server']) return String(errors['server']);
    if (errors['required'] || errors['blank']) return `${label} is required.`;
    if (errors['futureDate']) return 'Joining date cannot be in the future.';
    if (errors['emailFormat']) return 'Enter a valid email address, e.g. name@company.com.';
    if (errors['minlength']) return `${label} must be at least ${errors['minlength'].requiredLength} characters.`;
    if (errors['maxlength']) return `${label} must be at most ${errors['maxlength'].requiredLength} characters.`;
    return `${label} is invalid.`;
  }

  submit(): void {
    this.submitError.set(null);

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    if (this.saving()) return;

    const raw = this.form.getRawValue();
    const payload: EmployeePayload = {
      ...raw,
      name: raw.name.trim(),
      email: raw.email.trim().toLowerCase(),
      designation: raw.designation.trim(),
    };

    const id = this.employeeId();
    const request$ = id === null ? this.api.create(payload) : this.api.update(id, payload);

    this.saving.set(true);
    request$.subscribe({
      next: () => {
        this.saving.set(false);
        this.notify.success(id === null ? `${payload.name} was added.` : `${payload.name} was updated.`);
        void this.router.navigate(['/employees']);
      },
      error: (err) => {
        this.saving.set(false);
        this.showSaveError(err);
      },
    });
  }

  /** Map API errors onto the matching fields; fall back to a banner for everything else. */
  private showSaveError(err: unknown): void {
    const { status, message, fields } = parseApiError(err);
    let mapped = false;

    for (const [field, text] of Object.entries(fields)) {
      const control = this.form.get(field);
      if (control) {
        control.setErrors({ server: text });
        control.markAsTouched();
        mapped = true;
      }
    }

    // Duplicate email: 409 even if the backend did not name the field.
    if (!mapped && status === 409) {
      const email = this.form.controls.email;
      email.setErrors({ server: message });
      email.markAsTouched();
      mapped = true;
    }

    this.submitError.set(mapped ? 'Please fix the highlighted fields and try again.' : message);
  }
}
