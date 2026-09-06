import { Suspense, lazy } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"
import { Loader2 } from "lucide-react"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { RequireAuth } from "@/components/layout/RequireAuth"
import { CampaignDialog } from "@/components/campaigns/CampaignDialog"
import { HomePage } from "@/pages/HomePage"
import { LoginPage } from "@/pages/LoginPage"
import { ApiError } from "@/api/client"

// Code-split: recharts is heavy and only the analytics route needs it, so it
// stays out of the initial bundle.
const CampaignAnalyticsPage = lazy(() => import("@/pages/CampaignAnalyticsPage"))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Retrying a 401 or 404 just delays showing the user what happened.
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
          return false
        }
        return failureCount < 2
      },
      refetchOnWindowFocus: false,
    },
  },
})

function RouteFallback() {
  return (
    <div className="flex min-h-svh items-center justify-center">
      <Loader2 className="size-6 animate-spin text-muted-foreground" />
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider delayDuration={200}>
        <BrowserRouter>
          <Suspense fallback={<RouteFallback />}>
            <Routes>
              <Route path="/login" element={<LoginPage />} />

              <Route
                path="/"
                element={
                  <RequireAuth>
                    <HomePage />
                  </RequireAuth>
                }
              >
                {/* Nested so the dialog renders over the campaign list. */}
                <Route path="campaigns/:shortCode" element={<CampaignDialog />} />
              </Route>

              {/* Deliberately NOT nested under "/" - a full page, not a dialog
                  over the list. React Router ranks it above the dialog route. */}
              <Route
                path="/campaigns/:shortCode/analytics"
                element={
                  <RequireAuth>
                    <CampaignAnalyticsPage />
                  </RequireAuth>
                }
              />

              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </BrowserRouter>
      </TooltipProvider>

      <Toaster richColors position="top-center" />
    </QueryClientProvider>
  )
}
