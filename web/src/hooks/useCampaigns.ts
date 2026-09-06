import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { campaignsApi } from "@/api/campaigns"
import type { AddCampaignRequest, ShortenRequest, UpdateCampaignRequest } from "@/api/types"

export const CAMPAIGNS_KEY = ["campaigns"] as const
export const PAGE_SIZE = 10

export function useCampaigns(page: number) {
  return useQuery({
    queryKey: [...CAMPAIGNS_KEY, "list", page],
    queryFn: () => campaignsApi.list({ limit: PAGE_SIZE, offset: page * PAGE_SIZE }),
  })
}

export function useCampaign(shortCode: string | undefined) {
  return useQuery({
    queryKey: [...CAMPAIGNS_KEY, "detail", shortCode],
    queryFn: () => campaignsApi.get(shortCode!),
    enabled: Boolean(shortCode),
  })
}

export function useShortenUrl() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (body: ShortenRequest) => campaignsApi.shorten(body),
    onSuccess: (result) => {
      // Only a 201 changed the list; a 200 just reported what already existed.
      if (result.created) {
        queryClient.invalidateQueries({ queryKey: CAMPAIGNS_KEY })
      }
    },
  })
}

export function useAddCampaign() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ destinationId, body }: { destinationId: number; body: AddCampaignRequest }) =>
      campaignsApi.addToDestination(destinationId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CAMPAIGNS_KEY })
    },
  })
}

export function useUpdateCampaign() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ shortCode, patch }: { shortCode: string; patch: UpdateCampaignRequest }) =>
      campaignsApi.update(shortCode, patch),
    onSuccess: () => {
      // Refreshes both the list and the open detail view.
      queryClient.invalidateQueries({ queryKey: CAMPAIGNS_KEY })
    },
  })
}
