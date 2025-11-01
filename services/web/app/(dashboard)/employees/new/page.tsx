import { EmployeeForm } from '@/components/employees/EmployeeForm';

export default function NewEmployeePage() {
  return (
    <div className="container mx-auto py-8 max-w-2xl space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Create Employee</h1>
        <p className="text-muted-foreground mt-1">
          Add a new employee to your organization
        </p>
      </div>

      {/* Employee Form */}
      <EmployeeForm mode="create" />
    </div>
  );
}
