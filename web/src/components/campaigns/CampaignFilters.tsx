import { Search, X } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { STATUS_LABELS } from "./CampaignStatusBadge"
import type { CampaignStatus } from "@/api/types"

/**
 * Radix reserves "" as a Select value, so "any status" needs a sentinel of its
 * own rather than an empty option.
 */
const ANY_STATUS = "all"

const STATUS_OPTIONS = Object.keys(STATUS_LABELS) as CampaignStatus[]

type CampaignFiltersProps = {
  search: string
  onSearchChange: (search: string) => void
  status: CampaignStatus | null
  onStatusChange: (status: CampaignStatus | null) => void
}

export function CampaignFilters({
  search,
  onSearchChange,
  status,
  onStatusChange,
}: CampaignFiltersProps) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row">
      <div className="relative flex-1">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />

        <Input
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder="Search name, short code or destination"
          aria-label="Search campaigns"
          className="px-9"
        />

        {search && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Clear search"
            onClick={() => onSearchChange("")}
            className="absolute right-1 top-1/2 size-7 -translate-y-1/2"
          >
            <X className="size-4" />
          </Button>
        )}
      </div>

      <Select
        value={status ?? ANY_STATUS}
        onValueChange={(value) =>
          onStatusChange(value === ANY_STATUS ? null : (value as CampaignStatus))
        }
      >
        <SelectTrigger className="sm:w-[11rem]" aria-label="Filter by status">
          <SelectValue />
        </SelectTrigger>

        <SelectContent>
          <SelectItem value={ANY_STATUS}>All statuses</SelectItem>
          {STATUS_OPTIONS.map((option) => (
            <SelectItem key={option} value={option}>
              {STATUS_LABELS[option]}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
