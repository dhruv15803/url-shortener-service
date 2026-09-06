import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Info, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { CopyButton } from "@/components/common/CopyButton"
import { CampaignStatusBadge } from "./CampaignStatusBadge"
import { useAddCampaign } from "@/hooks/useCampaigns"
import { toApiInstant } from "@/lib/datetime"
import type { DestinationRef, ShortUrlRef } from "@/api/types"

const schema = z.object({
  campaign_name: z.string().trim().max(255, "Keep it under 255 characters"),
  starts_at: z.string(),
  expires_at: z.string(),
})

type FormValues = z.infer<typeof schema>

type ExistingDestinationPanelProps = {
  destination: DestinationRef
  shortUrls: ShortUrlRef[]
  onDone: () => void
}

/**
 * Shown when POST /api/urls answered 200 - the user had already shortened this
 * destination, so instead of silently creating a duplicate we show what exists
 * and offer to add another campaign to it.
 */
export function ExistingDestinationPanel({
  destination,
  shortUrls,
  onDone,
}: ExistingDestinationPanelProps) {
  const [added, setAdded] = useState<ShortUrlRef[]>([])
  const addCampaign = useAddCampaign()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { campaign_name: "", starts_at: "", expires_at: "" },
  })

  function onSubmit(values: FormValues) {
    addCampaign.mutate(
      {
        destinationId: destination.id,
        body: {
          campaign_name: values.campaign_name || null,
          starts_at: toApiInstant(values.starts_at),
          expires_at: toApiInstant(values.expires_at),
        },
      },
      {
        onSuccess: (shortUrl) => {
          setAdded((current) => [...current, shortUrl])
          form.reset({ campaign_name: "", starts_at: "", expires_at: "" })
          toast.success("Campaign added")
        },
        onError: (error) => toast.error(error.message),
      },
    )
  }

  const allShortUrls = [...shortUrls, ...added]

  return (
    <div className="space-y-4">
      <div className="flex items-start gap-2 rounded-lg border bg-muted/40 p-3">
        <Info className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0 space-y-1">
          <p className="text-sm font-medium">You've already shortened this URL</p>
          <p className="truncate text-sm text-muted-foreground">{destination.url}</p>
        </div>
        <Button variant="ghost" size="sm" onClick={onDone} className="ml-auto shrink-0">
          Shorten another
        </Button>
      </div>

      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground">
          Existing campaigns ({allShortUrls.length})
        </p>
        <ul className="divide-y rounded-lg border">
          {allShortUrls.map((shortUrl) => (
            <li key={shortUrl.id} className="flex items-center gap-3 p-3">
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">
                  {shortUrl.campaign_name ?? (
                    <span className="text-muted-foreground">Untitled</span>
                  )}
                </p>
                <a
                  href={shortUrl.url}
                  target="_blank"
                  rel="noreferrer"
                  className="font-mono text-xs text-muted-foreground underline-offset-4 hover:underline"
                >
                  {shortUrl.url}
                </a>
              </div>
              <CampaignStatusBadge status={shortUrl.status} />
              <CopyButton value={shortUrl.url} />
            </li>
          ))}
        </ul>
      </div>

      <Separator />

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <p className="text-sm font-medium">Create another campaign for this URL</p>

          <div className="grid items-start gap-4 sm:grid-cols-3">
            <FormField
              control={form.control}
              name="campaign_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Campaign name</FormLabel>
                  <FormControl>
                    <Input placeholder="Email campaign" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="starts_at"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Starts</FormLabel>
                  <FormControl>
                    <Input type="datetime-local" {...field} />
                  </FormControl>
                  <FormDescription>Leave empty to start now</FormDescription>
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="expires_at"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Expires</FormLabel>
                  <FormControl>
                    <Input type="datetime-local" {...field} />
                  </FormControl>
                  <FormDescription>Leave empty for never</FormDescription>
                </FormItem>
              )}
            />
          </div>

          <Button type="submit" disabled={addCampaign.isPending}>
            {addCampaign.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
            Add campaign
          </Button>
        </form>
      </Form>
    </div>
  )
}
