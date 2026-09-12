import { useNavigate } from "react-router-dom"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { CopyButton } from "@/components/common/CopyButton"
import { EmptyState, ErrorState, LoadingRows } from "@/components/common/States"
import { CampaignStatusBadge } from "./CampaignStatusBadge"
import { formatDateTime } from "@/lib/datetime"
import type { Campaign } from "@/api/types"

type CampaignTableProps = {
  campaigns: Campaign[]
  isLoading: boolean
  error: Error | null
  onRetry: () => void
  /** Changes the empty state: no matches is not the same as no campaigns. */
  isFiltered?: boolean
}

export function CampaignTable({
  campaigns,
  isLoading,
  error,
  onRetry,
  isFiltered = false,
}: CampaignTableProps) {
  const navigate = useNavigate()

  if (error) {
    return <ErrorState message={error.message} onRetry={onRetry} />
  }

  if (!isLoading && campaigns.length === 0) {
    return isFiltered ? (
      <EmptyState
        title="No campaigns match your filters"
        description="Try a different search term, or widen the status filter."
      />
    ) : (
      <EmptyState
        title="No campaigns yet"
        description="Shorten your first URL above and it will show up here."
      />
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Campaign</TableHead>
          <TableHead>Short link</TableHead>
          <TableHead className="hidden md:table-cell">Destination</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="hidden lg:table-cell">Expires</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {isLoading ? (
          <LoadingRows />
        ) : (
          campaigns.map((campaign) => (
            <TableRow
              key={campaign.id}
              onClick={() => navigate(`/campaigns/${campaign.code}`)}
              className="cursor-pointer"
            >
              <TableCell className="font-medium">
                {campaign.campaign_name ?? <span className="text-muted-foreground">Untitled</span>}
              </TableCell>

              <TableCell>
                <div className="flex items-center gap-1">
                  <a
                    href={campaign.short_url}
                    target="_blank"
                    rel="noreferrer"
                    onClick={(event) => event.stopPropagation()}
                    className="font-mono text-sm underline-offset-4 hover:underline"
                  >
                    /{campaign.code}
                  </a>
                  <span onClick={(event) => event.stopPropagation()}>
                    <CopyButton value={campaign.short_url} />
                  </span>
                </div>
              </TableCell>

              <TableCell className="hidden max-w-[22rem] truncate text-muted-foreground md:table-cell">
                {campaign.destination_url}
              </TableCell>

              <TableCell>
                <CampaignStatusBadge status={campaign.status} />
              </TableCell>

              <TableCell className="hidden text-muted-foreground lg:table-cell">
                {formatDateTime(campaign.expires_at, "Never")}
              </TableCell>
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}
