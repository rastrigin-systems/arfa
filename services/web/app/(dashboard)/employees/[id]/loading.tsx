import { Card, CardContent, CardHeader } from '@/components/ui/card';

export default function EmployeeDetailLoading() {
  return (
    <div className="space-y-6" role="status" aria-label="Loading employee details">
      {/* Header Skeleton */}
      <div className="flex items-center justify-between">
        <div>
          <div className="h-4 w-32 bg-muted rounded mb-2 animate-pulse" />
          <div className="h-8 w-64 bg-muted rounded animate-pulse" />
        </div>
        <div className="flex gap-2">
          <div className="h-10 w-20 bg-muted rounded animate-pulse" />
          <div className="h-10 w-20 bg-muted rounded animate-pulse" />
        </div>
      </div>

      {/* Basic Information Card Skeleton */}
      <Card>
        <CardHeader>
          <div className="h-6 w-48 bg-muted rounded animate-pulse" />
          <div className="h-4 w-64 bg-muted rounded animate-pulse mt-2" />
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[...Array(6)].map((_, i) => (
              <div key={i}>
                <div className="h-4 w-24 bg-muted rounded animate-pulse" />
                <div className="h-5 w-32 bg-muted rounded animate-pulse mt-2" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Teams Card Skeleton */}
      <Card>
        <CardHeader>
          <div className="h-6 w-32 bg-muted rounded animate-pulse" />
          <div className="h-4 w-56 bg-muted rounded animate-pulse mt-2" />
        </CardHeader>
        <CardContent>
          <div className="h-8 w-40 bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>

      {/* Agents Card Skeleton */}
      <Card>
        <CardHeader>
          <div className="h-6 w-32 bg-muted rounded animate-pulse" />
          <div className="h-4 w-64 bg-muted rounded animate-pulse mt-2" />
        </CardHeader>
        <CardContent className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="flex items-center justify-between p-3 border rounded-lg">
              <div>
                <div className="h-5 w-32 bg-muted rounded animate-pulse" />
                <div className="h-4 w-24 bg-muted rounded animate-pulse mt-2" />
              </div>
              <div className="h-6 w-16 bg-muted rounded-full animate-pulse" />
            </div>
          ))}
        </CardContent>
      </Card>

      {/* MCP Card Skeleton */}
      <Card>
        <CardHeader>
          <div className="h-6 w-40 bg-muted rounded animate-pulse" />
          <div className="h-4 w-72 bg-muted rounded animate-pulse mt-2" />
        </CardHeader>
        <CardContent>
          <div className="h-5 w-48 bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>

      <span className="sr-only">Loading employee details...</span>
    </div>
  );
}
