import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { useUpdateCampaign } from "@/hooks/useCampaigns"
import { toApiInstant, toInputValue } from "@/lib/datetime"
import type { Campaign, UpdateCampaignRequest } from "@/api/types"

const schema = z.object({
  name: z.string().trim().max(255, "Keep it under 255 characters"),
  enabled: z.boolean(),
  starts_at: z.string(),
  expires_at: z.string(),
})

type FormValues = z.infer<typeof schema>

function toFormValues(campaign: Campaign): FormValues {
  return {
    name: campaign.campaign_name ?? "",
    // Only `disabled` is a user decision; scheduled/expired are derived from
    // the dates, so anything other than disabled means "enabled".
    enabled: campaign.status !== "disabled",
    starts_at: toInputValue(campaign.starts_at),
    expires_at: toInputValue(campaign.expires_at),
  }
}

export function CampaignEditForm({
  campaign,
  onSaved,
}: {
  campaign: Campaign
  onSaved: () => void
}) {
  const update = useUpdateCampaign()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    values: toFormValues(campaign),
  })

  function onSubmit(values: FormValues) {
    // The api treats an omitted key as "leave alone" and an explicit null as
    // "clear", so emptied fields must be sent as null rather than dropped.
    const patch: UpdateCampaignRequest = {
      name: values.name === "" ? null : values.name,
      status: values.enabled ? "ACTIVE" : "DISABLED",
      starts_at: toApiInstant(values.starts_at),
      expires_at: toApiInstant(values.expires_at),
    }

    update.mutate(
      { shortCode: campaign.code, patch },
      {
        onSuccess: () => {
          toast.success("Campaign updated")
          onSaved()
        },
        onError: (error) => toast.error(error.message),
      },
    )
  }

  // Without this, a validation error on a field that renders no FormMessage
  // would block submit with no feedback at all.
  function onInvalid(errors: Record<string, { message?: string }>) {
    const first = Object.values(errors).find((error) => error?.message)
    toast.error(first?.message ?? "Please check the form and try again")
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit, onInvalid)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Campaign name</FormLabel>
              <FormControl>
                <Input placeholder="Untitled campaign" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="enabled"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between gap-4 rounded-lg border p-3">
              {/* min-w-0 lets the text shrink; without it the description's
                  min-content width pushes the switch outside the dialog. */}
              <div className="min-w-0 space-y-0.5">
                <FormLabel>Enabled</FormLabel>
                <FormDescription>
                  Turn off to stop the link redirecting, without losing its click history.
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  checked={field.value}
                  onCheckedChange={field.onChange}
                  className="shrink-0"
                />
              </FormControl>
            </FormItem>
          )}
        />

        <div className="grid items-start gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="starts_at"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Starts</FormLabel>
                <FormControl>
                  <Input type="datetime-local" {...field} />
                </FormControl>
                <FormDescription>Clear to start immediately</FormDescription>
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
                <FormDescription>Clear to never expire</FormDescription>
              </FormItem>
            )}
          />
        </div>

        <Button type="submit" disabled={update.isPending} className="w-full">
          {update.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
          Save changes
        </Button>
      </form>
    </Form>
  )
}
