import { Navigate, useSearchParams } from "react-router-dom"
import { AlertCircle, Link2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { authApi } from "@/api/auth"
import { useAuth } from "@/hooks/useAuth"

/** Error codes the api appends when the oauth callback fails. */
const ERROR_MESSAGES: Record<string, string> = {
  login_cancelled: "Sign-in was cancelled.",
  invalid_state: "That sign-in link expired. Please try again.",
  missing_code: "Google didn't return a sign-in code. Please try again.",
  login_failed: "Something went wrong signing you in. Please try again.",
}

export function LoginPage() {
  const [searchParams] = useSearchParams()
  const { isAuthenticated, isLoading } = useAuth()

  const errorCode = searchParams.get("error")
  const errorMessage = errorCode
    ? (ERROR_MESSAGES[errorCode] ?? ERROR_MESSAGES.login_failed)
    : null

  if (!isLoading && isAuthenticated) {
    return <Navigate to="/" replace />
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex size-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Link2 className="size-5" />
          </div>
          <CardTitle>Sign in to Shortly</CardTitle>
          <CardDescription>Shorten links and track their campaigns.</CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {errorMessage && (
            <div className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm">
              <AlertCircle className="mt-0.5 size-4 shrink-0 text-destructive" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* A full page navigation, not an xhr - google's consent screen has
              to be rendered by the browser. */}
          <Button asChild className="w-full">
            <a href={authApi.loginUrl()}>Continue with Google</a>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
