export default function EmployeesLoading() {
  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Page Header Skeleton */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <div className="h-9 w-48 bg-muted animate-pulse rounded-md" />
          <div className="h-5 w-64 bg-muted animate-pulse rounded-md" />
        </div>
        <div className="h-10 w-40 bg-muted animate-pulse rounded-md" />
      </div>

      {/* Filters Skeleton */}
      <div className="flex gap-4">
        <div className="h-10 w-80 bg-muted animate-pulse rounded-md" />
        <div className="h-10 w-[180px] bg-muted animate-pulse rounded-md" />
      </div>

      {/* Table Skeleton */}
      <div className="rounded-md border" role="status" aria-label="Loading employees">
        <div className="p-4">
          {/* Table Header */}
          <div className="flex gap-4 mb-4">
            <div className="h-6 w-24 bg-muted animate-pulse rounded-md" />
            <div className="h-6 w-32 bg-muted animate-pulse rounded-md" />
            <div className="h-6 w-20 bg-muted animate-pulse rounded-md" />
            <div className="h-6 w-20 bg-muted animate-pulse rounded-md" />
            <div className="h-6 w-20 bg-muted animate-pulse rounded-md" />
            <div className="h-6 w-24 bg-muted animate-pulse rounded-md" />
          </div>

          {/* Table Rows */}
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex gap-4 mb-3">
              <div className="h-10 w-32 bg-muted animate-pulse rounded-md" />
              <div className="h-10 w-48 bg-muted animate-pulse rounded-md" />
              <div className="h-10 w-24 bg-muted animate-pulse rounded-md" />
              <div className="h-10 w-20 bg-muted animate-pulse rounded-md" />
              <div className="h-10 w-24 bg-muted animate-pulse rounded-md" />
              <div className="h-10 w-20 bg-muted animate-pulse rounded-md" />
            </div>
          ))}
        </div>
        <span className="sr-only">Loading employees...</span>
      </div>
    </div>
  );
}
