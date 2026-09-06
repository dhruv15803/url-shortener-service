import { useNavigate, useParams } from "react-router-dom"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { CopyButton } from "@/components/common/CopyButton"
import { ErrorState, Field } from "@/components/common/States"
import { CampaignStatusBadge } from "./CampaignStatusBadge"
import { CampaignEditForm } from "./CampaignEditForm"
import { useCampaign } from "@/hooks/useCampaigns"
import { formatDateTime } from "@/lib/datetime"

/**
 * Backed by the /campaigns/:shortCode route rather than local state, so a
 * campaign can be linked to directly and the back button closes it.
 */
export function CampaignDialog() {
  const { shortCode } = useParams<{ shortCode: string }>()
  const navigate = useNavigate()
  const { data: campaign, isLoading, error, refetch } = useCampaign(shortCode)

  function close() {
    navigate("/")
  }

  return (
    <Dialog open onOpenChange={(open) => !open && close()}>
      {/* overflow-x-hidden matters: overflow-y-auto alone makes overflow-x
          compute to auto, so any unwrapped content adds a horizontal scrollbar. */}
      <DialogContent className="max-h-[90vh] overflow-x-hidden overflow-y-auto sm:max-w-lg">
        {isLoading && (
          <div className="space-y-4">
            <Skeleton className="h-6 w-40" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-40 w-full" />
          </div>
        )}

        {error && <ErrorState message={error.message} onRetry={() => refetch()} />}

        {campaign && (
          <>
            <DialogHeader className="min-w-0">
              {/* flex-wrap so a long name pushes the badge to the next line
                  instead of widening the dialog. */}
              <div className="flex flex-wrap items-center gap-2">
                <DialogTitle className="break-words">
                  {campaign.campaign_name ?? "Untitled campaign"}
                </DialogTitle>
                <CampaignStatusBadge status={campaign.status} />
              </div>
              {/* Destination urls are long, unbreakable strings - they have to
                  wrap, or their min-content width blows out the dialog. */}
              <DialogDescription className="break-all">
                {campaign.destination_url}
              </DialogDescription>
            </DialogHeader>

            <div className="flex items-center gap-2 rounded-lg border bg-muted/40 p-3">
              <a
                href={campaign.short_url}
                target="_blank"
                rel="noreferrer"
                className="min-w-0 flex-1 font-mono text-sm break-all underline-offset-4 hover:underline"
              >
                {campaign.short_url}
              </a>
              <CopyButton value={campaign.short_url} />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label="Created">{formatDateTime(campaign.created_at)}</Field>
              <Field label="Status">
                {/* Derived by the server from the dates below and the enabled toggle. */}
                <CampaignStatusBadge status={campaign.status} />
              </Field>
            </div>

            <Separator />

            <CampaignEditForm campaign={campaign} onSaved={close} />
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
