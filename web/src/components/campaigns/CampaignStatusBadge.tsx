import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { CampaignStatus } from "@/api/types"

/**
 * Status is computed by the server from the campaign's schedule and whether
 * it was disabled - it is never something the user picks directly.
 */
const STATUS_STYLES: Record<CampaignStatus, { label: string; className: string }> = {
  active: {
    label: "Active",
    className: "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950 dark:text-emerald-400",
  },
  scheduled: {
    label: "Scheduled",
    className: "border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-400",
  },
  expired: {
    label: "Expired",
    className: "border-neutral-200 bg-neutral-100 text-neutral-600 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-400",
  },
  disabled: {
    label: "Disabled",
    className: "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-400",
  },
}

export function CampaignStatusBadge({
  status,
  className,
}: {
  status: CampaignStatus
  className?: string
}) {
  const style = STATUS_STYLES[status]

  return (
    <Badge variant="outline" className={cn(style.className, className)}>
      {style.label}
    </Badge>
  )
}
