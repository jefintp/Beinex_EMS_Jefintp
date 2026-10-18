export type EmployeeStatus = 'active' | 'inactive';

/** Shape returned by the Go API (snake_case, matching the DB/JSON fields). */
export interface Employee {
  id: number;
  name: string;
  email: string;
  department: string;
  designation: string;
  /** Date only, `YYYY-MM-DD` (an ISO timestamp is tolerated too). */
  joining_date: string;
  status: EmployeeStatus;
}

/** Body sent on create / update. */
export type EmployeePayload = Omit<Employee, 'id'>;

/** Query parameters accepted by `GET /api/employees`. */
export interface EmployeeFilter {
  search?: string;
  department?: string;
  status?: string;
}

export const DEPARTMENTS: readonly string[] = [
  'Engineering',
  'Human Resources',
  'Finance',
  'Sales',
  'Marketing',
  'Operations',
  'Support',
  'IT',
];

export const STATUSES: readonly { value: EmployeeStatus; label: string }[] = [
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Inactive' },
];
