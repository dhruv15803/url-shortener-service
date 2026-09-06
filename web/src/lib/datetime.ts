import { format, parseISO } from "date-fns"

/**
 * The api speaks RFC3339 instants with offsets; <input type="datetime-local">
 * speaks local wall-clock with no zone at all. Both conversions live here so a
 * picker can't silently send a time shifted by the user's utc offset.
 */

const INPUT_FORMAT = "yyyy-MM-dd'T'HH:mm"

/** RFC3339 instant -> value for a datetime-local input, in local time. */
export function toInputValue(instant: string | null | undefined): string {
  if (!instant) return ""
  return format(parseISO(instant), INPUT_FORMAT)
}

/** datetime-local value (local time) -> RFC3339 instant for the api. */
export function toApiInstant(inputValue: string): string | null {
  if (!inputValue) return null

  // `new Date("2026-01-01T10:00")` is interpreted as local time, which is
  // exactly what the input meant; toISOString then converts to utc.
  const parsed = new Date(inputValue)
  if (Number.isNaN(parsed.getTime())) return null

  return parsed.toISOString()
}

/** Human-readable date for display, or a placeholder when unset. */
export function formatDateTime(instant: string | null | undefined, fallback = "—"): string {
  if (!instant) return fallback
  return format(parseISO(instant), "d MMM yyyy, HH:mm")
}
