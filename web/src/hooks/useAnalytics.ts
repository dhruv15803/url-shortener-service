import { useQueries, useQuery } from "@tanstack/react-query"
import { analyticsApi } from "@/api/analytics"
import type { AnalyticsRange, ClickDimension } from "@/api/types"

export const ANALYTICS_KEY = ["analytics"] as const

// The range is part of the key, so switching presets refetches rather than
// showing the previous range's numbers.
function rangeKey(range: AnalyticsRange) {
  return [range.start.toISOString(), range.end.toISOString()]
}

export function useAnalyticsOverview(shortCode: string | undefined, range: AnalyticsRange) {
  return useQuery({
    queryKey: [...ANALYTICS_KEY, shortCode, "overview", ...rangeKey(range)],
    queryFn: () => analyticsApi.overview(shortCode!, range),
    enabled: Boolean(shortCode),
  })
}

/**
 * One query per dimension, run in parallel. The api serves a single dimension
 * per request, but each is the same indexed query and cached server-side, so
 * fetching them together is cheap and the whole dashboard fills at once.
 */
export function useAnalyticsBreakdowns(
  shortCode: string | undefined,
  dimensions: ClickDimension[],
  range: AnalyticsRange,
) {
  return useQueries({
    queries: dimensions.map((dimension) => ({
      queryKey: [...ANALYTICS_KEY, shortCode, "breakdown", dimension, ...rangeKey(range)],
      queryFn: () => analyticsApi.breakdown(shortCode!, dimension, range),
      enabled: Boolean(shortCode),
    })),
  })
}
