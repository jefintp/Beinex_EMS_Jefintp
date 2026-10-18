import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'employees' },
  {
    path: 'employees',
    title: 'Employees',
    loadComponent: () =>
      import('./employees/employee-list/employee-list.component').then(
        (m) => m.EmployeeListComponent,
      ),
  },
  {
    path: 'employees/new',
    title: 'Add employee',
    loadComponent: () =>
      import('./employees/employee-form/employee-form.component').then(
        (m) => m.EmployeeFormComponent,
      ),
  },
  {
    path: 'employees/:id/edit',
    title: 'Edit employee',
    loadComponent: () =>
      import('./employees/employee-form/employee-form.component').then(
        (m) => m.EmployeeFormComponent,
      ),
  },
  { path: '**', redirectTo: 'employees' },
];
