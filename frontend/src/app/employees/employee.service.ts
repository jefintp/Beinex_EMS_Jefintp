import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { Employee, EmployeeFilter, EmployeePayload } from './employee.model';

/** The Go API wraps the list as `{ "data": [...], "total": n }`. */
interface EmployeeListResponse {
  data: Employee[] | null;
  total: number;
}

@Injectable({ providedIn: 'root' })
export class EmployeeService {
  private readonly http = inject(HttpClient);

  /** Proxied to the Go backend in dev (see proxy.conf.json). */
  private readonly baseUrl = '/api/v1/employees';

  list(filter: EmployeeFilter = {}): Observable<Employee[]> {
    let params = new HttpParams();
    const search = filter.search?.trim();
    if (search) params = params.set('search', search);
    if (filter.department) params = params.set('department', filter.department);
    if (filter.status) params = params.set('status', filter.status);

    // Unwrap `{ data, total }`; normalise a missing/null list to an empty array.
    return this.http
      .get<EmployeeListResponse | null>(this.baseUrl, { params })
      .pipe(map((res) => res?.data ?? []));
  }

  get(id: number): Observable<Employee> {
    return this.http.get<Employee>(`${this.baseUrl}/${id}`);
  }

  create(payload: EmployeePayload): Observable<Employee> {
    return this.http.post<Employee>(this.baseUrl, payload);
  }

  update(id: number, payload: EmployeePayload): Observable<Employee> {
    return this.http.put<Employee>(`${this.baseUrl}/${id}`, payload);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
