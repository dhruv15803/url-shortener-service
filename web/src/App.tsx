import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"
import { Toaster } from "@/components/ui/sonner"
import { RequireAuth } from "@/components/layout/RequireAuth"
import { CampaignDialog } from "@/components/campaigns/CampaignDialog"
import { HomePage } from "@/pages/HomePage"
import { LoginPage } from "@/pages/LoginPage"
import { ApiError } from "@/api/client"

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

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
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

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>

      <Toaster richColors position="top-center" />
    </QueryClientProvider>
  )
}
