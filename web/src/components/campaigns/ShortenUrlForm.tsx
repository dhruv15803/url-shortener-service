import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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
import { ExistingDestinationPanel } from "./ExistingDestinationPanel"
import { useShortenUrl } from "@/hooks/useCampaigns"
import { toApiInstant } from "@/lib/datetime"
import type { ShortenResult } from "@/api/types"

const schema = z.object({
  destination_url: z.string().trim().min(1, "Enter a URL to shorten"),
  campaign_name: z.string().trim().max(255, "Keep it under 255 characters"),
  starts_at: z.string(),
  expires_at: z.string(),
})

type FormValues = z.infer<typeof schema>

const EMPTY: FormValues = {
  destination_url: "",
  campaign_name: "",
  starts_at: "",
  expires_at: "",
}

export function ShortenUrlForm() {
  const [result, setResult] = useState<ShortenResult | null>(null)
  const shorten = useShortenUrl()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: EMPTY,
  })

  function onSubmit(values: FormValues) {
    shorten.mutate(
      {
        destination_url: values.destination_url,
        campaign_name: values.campaign_name || null,
        starts_at: toApiInstant(values.starts_at),
        expires_at: toApiInstant(values.expires_at),
      },
      {
        onSuccess: (data) => {
          setResult(data)
          if (data.created) {
            form.reset(EMPTY)
            toast.success("Short link created")
          }
        },
        onError: (error) => toast.error(error.message),
      },
    )
  }

  // The destination already existed, so nothing was created - switch to the
  // panel that lets the user add another campaign to it instead.
  if (result && !result.created) {
    return (
      <ExistingDestinationPanel
        destination={result.destination}
        shortUrls={result.shortUrls}
        onDone={() => {
          setResult(null)
          form.reset(EMPTY)
        }}
      />
    )
  }

  return (
    <div className="space-y-4">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <FormField
            control={form.control}
            name="destination_url"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Destination URL</FormLabel>
                <FormControl>
                  <Input placeholder="https://example.com/product/123" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          {/* items-start stops the shorter field (no description) from being
              stretched to the taller ones, which would push its label down. */}
          <div className="grid items-start gap-4 sm:grid-cols-3">
            <FormField
              control={form.control}
              name="campaign_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Campaign name</FormLabel>
                  <FormControl>
                    <Input placeholder="Instagram launch" {...field} />
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

          <Button type="submit" disabled={shorten.isPending}>
            {shorten.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
            Shorten URL
          </Button>
        </form>
      </Form>

      {result?.created && (
        <div className="flex items-center justify-between gap-3 rounded-lg border bg-muted/40 p-3">
          <div className="min-w-0">
            <p className="text-xs text-muted-foreground">Your short link</p>
            <a
              href={result.shortUrl.url}
              target="_blank"
              rel="noreferrer"
              className="font-mono text-sm underline-offset-4 hover:underline"
            >
              {result.shortUrl.url}
            </a>
          </div>
          <CopyButton value={result.shortUrl.url} />
        </div>
      )}
    </div>
  )
}
