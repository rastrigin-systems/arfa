export default function PoliciesLoading() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <div className="h-10 w-48 bg-muted animate-pulse rounded-md" />
        <div className="h-5 w-96 mt-2 bg-muted animate-pulse rounded-md" />
      </div>

      {/* Search and filter */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex flex-col sm:flex-row gap-4 flex-1">
          <div className="h-10 w-full max-w-md bg-muted animate-pulse rounded-md" />
          <div className="h-10 w-40 bg-muted animate-pulse rounded-md" />
        </div>
        <div className="h-10 w-32 bg-muted animate-pulse rounded-md" />
      </div>

      {/* Policy cards skeleton grid */}
      <div
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        role="status"
        aria-label="Loading policies"
      >
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="rounded-lg border p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="h-6 w-24 bg-muted animate-pulse rounded-md" />
              <div className="h-6 w-16 bg-muted animate-pulse rounded-md" />
            </div>
            <div className="h-4 w-32 bg-muted animate-pulse rounded-md" />
            <div className="h-4 w-full bg-muted animate-pulse rounded-md" />
            <div className="h-4 w-24 bg-muted animate-pulse rounded-md" />
            <div className="flex gap-2 mt-4">
              <div className="h-8 flex-1 bg-muted animate-pulse rounded-md" />
              <div className="h-8 flex-1 bg-muted animate-pulse rounded-md" />
            </div>
          </div>
        ))}
        <span className="sr-only">Loading policies...</span>
      </div>
    </div>
  );
}
