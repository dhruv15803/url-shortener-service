import { addDays, addHours, differenceInHours, format, startOfDay, startOfHour } from "date-fns"
import type { AnalyticsRange, SeriesPoint } from "@/api/types"

/** Matches the server's rule: hourly for ranges up to 2 days, daily beyond. */
export type Granularity = "hour" | "day"

const HOURLY_THRESHOLD_HOURS = 48

export function granularityFor(range: AnalyticsRange): Granularity {
  return differenceInHours(range.end, range.start) <= HOURLY_THRESHOLD_HOURS ? "hour" : "day"
}

export type ChartPoint = {
  bucket: Date
  clicks: number
  label: string
}

/**
 * The api groups by time bucket, so buckets with zero clicks are absent from
 * the response entirely. Charting that directly would draw a straight line
 * across quiet periods and overstate how steady a campaign was, so every
 * missing bucket is filled with zero before plotting.
 */
export function densifySeries(
  points: SeriesPoint[],
  range: AnalyticsRange,
  granularity: Granularity,
): ChartPoint[] {
  const step = granularity === "hour" ? addHours : addDays
  const floor = granularity === "hour" ? startOfHour : startOfDay

  const clicksByBucket = new Map<number, number>()
  for (const point of points) {
    clicksByBucket.set(floor(new Date(point.bucket)).getTime(), point.clicks)
  }

  const filled: ChartPoint[] = []
  const last = floor(range.end).getTime()

  // Guard against a pathological range producing an unbounded loop.
  for (
    let cursor = floor(range.start), guard = 0;
    cursor.getTime() <= last && guard < 2000;
    cursor = step(cursor, 1), guard++
  ) {
    filled.push({
      bucket: cursor,
      clicks: clicksByBucket.get(cursor.getTime()) ?? 0,
      label: formatBucket(cursor, granularity),
    })
  }

  return filled
}

export function formatBucket(at: Date, granularity: Granularity): string {
  return granularity === "hour" ? format(at, "HH:mm") : format(at, "d MMM")
}

export function formatBucketFull(at: Date, granularity: Granularity): string {
  return granularity === "hour" ? format(at, "d MMM, HH:mm") : format(at, "d MMM yyyy")
}

/** Readable labels for the dimension cards. */
export const DIMENSION_LABELS: Record<string, string> = {
  country: "Country",
  city: "City",
  region: "Region",
  device: "Device",
  browser: "Browser",
  os: "Operating system",
}

/**
 * The api folds rows with no value into this bucket. It is often the largest
 * one - private ips don't geolocate and many user agents don't parse - so it
 * gets called out rather than shown as if it were a real value.
 */
export const UNKNOWN_VALUE = "unknown"

export function formatCompact(value: number): string {
  return new Intl.NumberFormat(undefined, { notation: "compact", maximumFractionDigits: 1 }).format(
    value,
  )
}
