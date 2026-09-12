import { useState } from "react"
import { Outlet } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { AppHeader } from "@/components/layout/AppHeader"
import { ShortenUrlForm } from "@/components/campaigns/ShortenUrlForm"
import { CampaignFilters } from "@/components/campaigns/CampaignFilters"
import { CampaignTable } from "@/components/campaigns/CampaignTable"
import { PAGE_SIZE, useCampaigns } from "@/hooks/useCampaigns"
import { useDebouncedValue } from "@/hooks/useDebouncedValue"
import type { CampaignStatus } from "@/api/types"

export function HomePage() {
  const [page, setPage] = useState(0)
  const [search, setSearch] = useState("")
  const [status, setStatus] = useState<CampaignStatus | null>(null)

  // The input stays on the raw value so typing feels immediate; only the
  // settled value reaches the query.
  const debouncedSearch = useDebouncedValue(search)
  const isFiltered = debouncedSearch !== "" || status !== null

  const { data, isLoading, error, refetch } = useCampaigns(page, {
    search: debouncedSearch,
    status,
  })

  const total = data?.total ?? 0
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const isLastPage = page >= pageCount - 1

  // Filtering from page 3 would otherwise request an offset past the end of a
  // narrower result and render an empty table.
  function handleSearchChange(value: string) {
    setSearch(value)
    setPage(0)
  }

  function handleStatusChange(value: CampaignStatus | null) {
    setStatus(value)
    setPage(0)
  }

  return (
    <div className="min-h-svh bg-muted/20">
      <AppHeader />

      <main className="mx-auto max-w-5xl space-y-6 px-4 py-8">
        <Card>
          <CardHeader>
            <CardTitle>Shorten a URL</CardTitle>
            <CardDescription>
              Paste a destination URL to create a short link and its first campaign.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ShortenUrlForm />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Your campaigns</CardTitle>
            <CardDescription>
              {isFiltered
                ? `${total} matching campaign${total === 1 ? "" : "s"}`
                : total > 0
                  ? `${total} campaign${total === 1 ? "" : "s"}`
                  : "Every short link you create shows up here."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <CampaignFilters
              search={search}
              onSearchChange={handleSearchChange}
              status={status}
              onStatusChange={handleStatusChange}
            />

            <CampaignTable
              campaigns={data?.campaigns ?? []}
              isLoading={isLoading}
              error={error}
              onRetry={() => refetch()}
              isFiltered={isFiltered}
            />

            {total > PAGE_SIZE && (
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Page {page + 1} of {pageCount}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page === 0}
                    onClick={() => setPage((current) => current - 1)}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={isLastPage}
                    onClick={() => setPage((current) => current + 1)}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </main>

      {/* The campaign dialog renders here, over the list. */}
      <Outlet />
    </div>
  )
}
