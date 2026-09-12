import { useMemo, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { ArrowLeft, Pencil } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { AppHeader } from "@/components/layout/AppHeader"
import { CopyButton } from "@/components/common/CopyButton"
import { EmptyState, ErrorState } from "@/components/common/States"
import { CampaignStatusBadge } from "@/components/campaigns/CampaignStatusBadge"
import {
  AnalyticsRangePicker,
  DEFAULT_PRESET_ID,
  RANGE_PRESETS,
} from "@/components/analytics/AnalyticsRangePicker"
import { BreakdownTable } from "@/components/analytics/BreakdownTable"
import { ClicksTrendChart } from "@/components/analytics/ClicksTrendChart"
import { useCampaign } from "@/hooks/useCampaigns"
import {
  BREAKDOWN_PAGE_SIZE,
  useAnalyticsBreakdown,
  useAnalyticsOverview,
} from "@/hooks/useAnalytics"
import { densifySeries, granularityFor } from "@/lib/analytics"
import { ApiError } from "@/api/client"
import type { AnalyticsRange, ClickDimension } from "@/api/types"

const defaultRange = () =>
  RANGE_PRESETS.find((preset) => preset.id === DEFAULT_PRESET_ID)!.build()

export default function CampaignAnalyticsPage() {
  const { shortCode } = useParams<{ shortCode: string }>()

  const [presetId, setPresetId] = useState<string | null>(DEFAULT_PRESET_ID)
  const [range, setRange] = useState<AnalyticsRange>(defaultRange)
  const [dimension, setDimension] = useState<ClickDimension | null>(null)
  const [page, setPage] = useState(1)

  const campaign = useCampaign(shortCode)
  const overview = useAnalyticsOverview(shortCode, range)
  const breakdown = useAnalyticsBreakdown(shortCode, dimension, range, page)

  const granularity = granularityFor(range)
  const points = useMemo(
    () => densifySeries(overview.data?.series ?? [], range, granularity),
    [overview.data?.series, range, granularity],
  )

  const totalClicks = overview.data?.total_clicks ?? 0
  const notFound = campaign.error instanceof ApiError && campaign.error.status === 404

  return (
    <div className="min-h-svh bg-muted/20">
      <AppHeader />

      <main className="mx-auto max-w-5xl space-y-6 px-4 py-8">
        <Button asChild variant="ghost" size="sm" className="-ml-2 gap-2">
          <Link to="/">
            <ArrowLeft className="size-4" />
            All campaigns
          </Link>
        </Button>

        {notFound ? (
          <ErrorState message="That campaign doesn't exist, or isn't yours." />
        ) : (
          <>
            <CampaignHeader shortCode={shortCode} campaign={campaign} />

            <div className="flex flex-wrap items-center justify-between gap-3">
              <AnalyticsRangePicker
                presetId={presetId}
                range={range}
                onChange={(nextRange, nextPreset) => {
                  setRange(nextRange)
                  setPresetId(nextPreset)
                  // A narrower range can have fewer pages than we're on.
                  setPage(1)
                }}
              />
            </div>

            <Card>
              <CardHeader>
                <CardDescription>Total clicks</CardDescription>
                <CardTitle className="text-3xl tabular-nums">
                  {overview.isLoading ? (
                    <Skeleton className="h-9 w-24" />
                  ) : (
                    totalClicks.toLocaleString()
                  )}
                </CardTitle>
              </CardHeader>

              <CardContent>
                {overview.isLoading && <Skeleton className="h-[260px] w-full" />}

                {!overview.isLoading && overview.error && (
                  <ErrorState message={overview.error.message} onRetry={() => overview.refetch()} />
                )}

                {!overview.isLoading && !overview.error && totalClicks === 0 && (
                  <EmptyState
                    title="No clicks in this range"
                    description="Nobody has followed this link during the selected period. Try a wider range."
                  />
                )}

                {!overview.isLoading && !overview.error && totalClicks > 0 && (
                  <ClicksTrendChart points={points} granularity={granularity} />
                )}
              </CardContent>
            </Card>

            <BreakdownTable
              dimension={dimension}
              onDimensionChange={(next) => {
                setDimension(next)
                // Dimensions have different page counts; staying on page 3
                // while switching to a 3-row dimension shows an empty table.
                setPage(1)
              }}
              rows={breakdown.data?.data ?? []}
              totalClicks={totalClicks}
              totalGroups={breakdown.data?.total_groups ?? 0}
              page={page}
              pageSize={BREAKDOWN_PAGE_SIZE}
              onPageChange={setPage}
              isLoading={dimension !== null && breakdown.isLoading}
              error={(breakdown.error as Error | null) ?? null}
            />
          </>
        )}
      </main>
    </div>
  )
}

function CampaignHeader({
  shortCode,
  campaign,
}: {
  shortCode: string | undefined
  campaign: ReturnType<typeof useCampaign>
}) {
  if (campaign.isLoading) {
    return <Skeleton className="h-20 w-full" />
  }

  if (!campaign.data) {
    return null
  }

  const data = campaign.data

  return (
    <div className="flex flex-wrap items-start justify-between gap-4">
      <div className="min-w-0 space-y-1">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-xl font-semibold break-words">
            {data.campaign_name ?? "Untitled campaign"}
          </h1>
          <CampaignStatusBadge status={data.status} />
        </div>

        <p className="text-sm break-all text-muted-foreground">{data.destination_url}</p>

        <div className="flex items-center gap-1">
          <a
            href={data.short_url}
            target="_blank"
            rel="noreferrer"
            className="font-mono text-sm break-all underline-offset-4 hover:underline"
          >
            {data.short_url}
          </a>
          <CopyButton value={data.short_url} />
        </div>
      </div>

      <Button asChild variant="outline" size="sm" className="shrink-0 gap-2">
        <Link to={`/campaigns/${shortCode}`}>
          <Pencil className="size-4" />
          Edit
        </Link>
      </Button>
    </div>
  )
}
