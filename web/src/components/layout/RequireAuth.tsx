import type { ReactNode } from "react"
import { Navigate } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { useAuth } from "@/hooks/useAuth"
import { ErrorState } from "@/components/common/States"

export function RequireAuth({ children }: { children: ReactNode }) {
  const { isLoading, isAuthenticated, isUnauthenticated, error } = useAuth()

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (isUnauthenticated) {
    return <Navigate to="/login" replace />
  }

  // A genuine failure (server down, cors) is distinct from being signed out,
  // and shouldn't bounce the user to a login page that also won't work.
  if (error) {
    return (
      <div className="flex min-h-svh items-center justify-center p-4">
        <ErrorState message={error.message} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}
