import { useQuery } from "@tanstack/react-query"
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

/** Rows per page in the breakdown table. */
export const BREAKDOWN_PAGE_SIZE = 20

/**
 * One dimension at a time, paged. `enabled` is what keeps the ungrouped
 * default free of requests: the overview query already carries total_clicks,
 * so nothing is fetched until a dimension is actually picked.
 */
export function useAnalyticsBreakdown(
  shortCode: string | undefined,
  dimension: ClickDimension | null,
  range: AnalyticsRange,
  page: number,
) {
  return useQuery({
    queryKey: [...ANALYTICS_KEY, shortCode, "breakdown", dimension, page, ...rangeKey(range)],
    queryFn: () =>
      analyticsApi.breakdown(shortCode!, dimension!, range, {
        limit: BREAKDOWN_PAGE_SIZE,
        offset: BREAKDOWN_PAGE_SIZE * (page - 1),
      }),
    enabled: Boolean(shortCode) && dimension !== null,
  })
}
