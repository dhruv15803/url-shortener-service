import { useState } from "react"
import type { DateRange } from "react-day-picker"
import { CalendarIcon } from "lucide-react"
import { endOfDay, format, startOfDay, subDays, subHours } from "date-fns"

import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"
import type { AnalyticsRange } from "@/api/types"

/** The api rejects anything longer; enforced here so the user finds out immediately. */
export const MAX_RANGE_DAYS = 90

type Preset = {
  id: string
  label: string
  build: () => AnalyticsRange
}

export const RANGE_PRESETS: Preset[] = [
  { id: "24h", label: "24h", build: () => ({ start: subHours(new Date(), 24), end: new Date() }) },
  { id: "7d", label: "7 days", build: () => ({ start: subDays(new Date(), 7), end: new Date() }) },
  { id: "30d", label: "30 days", build: () => ({ start: subDays(new Date(), 30), end: new Date() }) },
  { id: "90d", label: "90 days", build: () => ({ start: subDays(new Date(), 89), end: new Date() }) },
]

export const DEFAULT_PRESET_ID = "7d"

type AnalyticsRangePickerProps = {
  presetId: string | null
  range: AnalyticsRange
  onChange: (range: AnalyticsRange, presetId: string | null) => void
}

export function AnalyticsRangePicker({ presetId, range, onChange }: AnalyticsRangePickerProps) {
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState<DateRange | undefined>({ from: range.start, to: range.end })
  const [error, setError] = useState<string | null>(null)

  function applyCustom() {
    if (!draft?.from) return

    // A single clicked day means that whole day.
    const start = startOfDay(draft.from)
    const end = endOfDay(draft.to ?? draft.from)

    const days = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)
    if (days > MAX_RANGE_DAYS) {
      setError(`Pick a range of ${MAX_RANGE_DAYS} days or fewer`)
      return
    }

    setError(null)
    onChange({ start, end }, null)
    setOpen(false)
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      {RANGE_PRESETS.map((preset) => (
        <Button
          key={preset.id}
          variant={presetId === preset.id ? "default" : "outline"}
          size="sm"
          onClick={() => onChange(preset.build(), preset.id)}
        >
          {preset.label}
        </Button>
      ))}

      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant={presetId === null ? "default" : "outline"}
            size="sm"
            className={cn("gap-2", presetId === null && "font-medium")}
          >
            <CalendarIcon className="size-4" />
            {presetId === null
              ? `${format(range.start, "d MMM")} – ${format(range.end, "d MMM")}`
              : "Custom"}
          </Button>
        </PopoverTrigger>

        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="range"
            defaultMonth={range.start}
            selected={draft}
            onSelect={(next) => {
              setDraft(next)
              setError(null)
            }}
            disabled={{ after: new Date() }}
            numberOfMonths={1}
          />
          <div className="space-y-2 border-t p-3">
            {error && <p className="text-xs text-destructive">{error}</p>}
            <div className="flex justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button size="sm" disabled={!draft?.from} onClick={applyCustom}>
                Apply
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
