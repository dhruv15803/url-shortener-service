import { client } from "./client"
import type { AnalyticsRange, ClickAnalytics, ClickDimension } from "./types"

function rangeParams(range: AnalyticsRange) {
  return {
    start_ts: range.start.toISOString(),
    end_ts: range.end.toISOString(),
  }
}

export const analyticsApi = {
  /** No group_by: total clicks plus the bucketed series for the trend line. */
  async overview(shortCode: string, range: AnalyticsRange): Promise<ClickAnalytics> {
    const { data } = await client.get<ClickAnalytics>(`/api/urls/${shortCode}/clicks`, {
      params: rangeParams(range),
    })
    return data
  },

  /** One dimension's breakdown, busiest first. */
  async breakdown(
    shortCode: string,
    dimension: ClickDimension,
    range: AnalyticsRange,
    limit = 10,
  ): Promise<ClickAnalytics> {
    const { data } = await client.get<ClickAnalytics>(`/api/urls/${shortCode}/clicks`, {
      params: { ...rangeParams(range), group_by: dimension, limit },
    })
    return data
  },
}
