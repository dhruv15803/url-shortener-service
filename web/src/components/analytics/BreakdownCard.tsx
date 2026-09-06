import { HelpCircle } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { DIMENSION_LABELS, UNKNOWN_VALUE } from "@/lib/analytics"
import { cn } from "@/lib/utils"
import type { BreakdownRow, ClickDimension } from "@/api/types"

type BreakdownCardProps = {
  dimension: ClickDimension
  rows: BreakdownRow[]
  isLoading: boolean
  error: Error | null
}

export function BreakdownCard({ dimension, rows, isLoading, error }: BreakdownCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{DIMENSION_LABELS[dimension] ?? dimension}</CardTitle>
      </CardHeader>

      <CardContent>
        {isLoading && (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-8 w-full" />
            ))}
          </div>
        )}

        {!isLoading && error && <p className="py-6 text-sm text-muted-foreground">{error.message}</p>}

        {!isLoading && !error && rows.length === 0 && (
          <p className="py-6 text-sm text-muted-foreground">No data in this range.</p>
        )}

        {!isLoading && !error && rows.length > 0 && (
          <ul className="space-y-3">
            {rows.map((row) => (
              <BreakdownRowItem key={row.value} row={row} />
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}

function BreakdownRowItem({ row }: { row: BreakdownRow }) {
  const isUnknown = row.value === UNKNOWN_VALUE

  return (
    <li className="space-y-1.5">
      <div className="flex items-center justify-between gap-2 text-sm">
        <span className={cn("flex min-w-0 items-center gap-1", isUnknown && "text-muted-foreground")}>
          <span className="truncate">{isUnknown ? "Unknown" : row.value}</span>

          {/* This bucket is often the largest one, so say why rather than
              letting it look like a bug. */}
          {isUnknown && (
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="size-3.5 shrink-0 opacity-70" />
              </TooltipTrigger>
              <TooltipContent className="max-w-[16rem]">
                Clicks we couldn't attribute — private or local IPs have no location, and some
                user agents can't be identified.
              </TooltipContent>
            </Tooltip>
          )}
        </span>

        <span className="shrink-0 tabular-nums text-muted-foreground">
          {row.clicks.toLocaleString()} · {row.percentage}%
        </span>
      </div>

      <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={cn("h-full rounded-full", isUnknown ? "bg-muted-foreground/40" : "bg-primary")}
          style={{ width: `${Math.min(100, row.percentage)}%` }}
        />
      </div>
    </li>
  )
}
