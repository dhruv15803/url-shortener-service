import { HelpCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { EmptyState, ErrorState } from "@/components/common/States"
import { DIMENSION_LABELS, UNKNOWN_VALUE } from "@/lib/analytics"
import { cn } from "@/lib/utils"
import type { BreakdownRow, ClickDimension } from "@/api/types"

/** null is the ungrouped view: a single row for the campaign total. */
export const BREAKDOWN_DIMENSIONS: (ClickDimension | null)[] = [
  null,
  "country",
  "city",
  "region",
  "device",
  "browser",
  "os",
]

const SHORT_LABELS: Record<string, string> = {
  ...DIMENSION_LABELS,
  os: "OS",
}

type BreakdownTableProps = {
  dimension: ClickDimension | null
  onDimensionChange: (dimension: ClickDimension | null) => void
  rows: BreakdownRow[]
  totalClicks: number
  totalGroups: number
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  isLoading: boolean
  error: Error | null
}

export function BreakdownTable({
  dimension,
  onDimensionChange,
  rows,
  totalClicks,
  totalGroups,
  page,
  pageSize,
  onPageChange,
  isLoading,
  error,
}: BreakdownTableProps) {
  // Ungrouped: derived from the overview total, so no request and no paging.
  const isUngrouped = dimension === null
  const displayRows: BreakdownRow[] = isUngrouped
    ? [{ value: "All clicks", clicks: totalClicks, percentage: 100 }]
    : rows

  const pageCount = Math.max(1, Math.ceil(totalGroups / pageSize))
  const showPagination = !isUngrouped && totalGroups > pageSize

  return (
    <Card>
      <CardHeader className="gap-3">
        <CardTitle className="text-base">Clicks breakdown</CardTitle>

        <div className="flex flex-wrap gap-2">
          {BREAKDOWN_DIMENSIONS.map((option) => (
            <Button
              key={option ?? "all"}
              variant={option === dimension ? "default" : "outline"}
              size="sm"
              onClick={() => onDimensionChange(option)}
            >
              {option === null ? "All" : (SHORT_LABELS[option] ?? option)}
            </Button>
          ))}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {error ? (
          <ErrorState message={error.message} />
        ) : totalClicks === 0 && !isLoading ? (
          <EmptyState
            title="No clicks in this range"
            description="Nobody has followed this link during the selected period. Try a wider range."
          />
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{isUngrouped ? "All" : (SHORT_LABELS[dimension] ?? dimension)}</TableHead>
                  <TableHead className="w-[6rem] text-right">Clicks</TableHead>
                  <TableHead className="w-[10rem]">Percentage</TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                {isLoading
                  ? Array.from({ length: 5 }).map((_, index) => (
                      <TableRow key={index}>
                        <TableCell colSpan={3}>
                          <Skeleton className="h-5 w-full" />
                        </TableCell>
                      </TableRow>
                    ))
                  : displayRows.map((row) => <BreakdownTableRow key={row.value} row={row} />)}
              </TableBody>
            </Table>

            {showPagination && (
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Page {page} of {pageCount}
                  <span className="ml-2 hidden sm:inline">
                    ({totalGroups.toLocaleString()} values)
                  </span>
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => onPageChange(page - 1)}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= pageCount}
                    onClick={() => onPageChange(page + 1)}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

function BreakdownTableRow({ row }: { row: BreakdownRow }) {
  const isUnknown = row.value === UNKNOWN_VALUE

  return (
    <TableRow>
      <TableCell className={cn("font-medium", isUnknown && "font-normal text-muted-foreground")}>
        <span className="flex min-w-0 items-center gap-1">
          <span className="truncate">{isUnknown ? "Unknown" : row.value}</span>

          {/* Often the largest bucket, so say why rather than letting it look
              like a bug. */}
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
      </TableCell>

      <TableCell className="text-right tabular-nums">{row.clicks.toLocaleString()}</TableCell>

      <TableCell>
        <div className="flex items-center gap-2">
          <span className="w-[3.25rem] shrink-0 tabular-nums text-muted-foreground">
            {row.percentage}%
          </span>
          <div className="h-1.5 min-w-0 flex-1 overflow-hidden rounded-full bg-muted">
            <div
              className={cn(
                "h-full rounded-full",
                isUnknown ? "bg-muted-foreground/40" : "bg-primary",
              )}
              style={{ width: `${Math.min(100, row.percentage)}%` }}
            />
          </div>
        </div>
      </TableCell>
    </TableRow>
  )
}
