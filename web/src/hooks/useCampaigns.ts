import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { campaignsApi } from "@/api/campaigns"
import type {
  AddCampaignRequest,
  CampaignStatus,
  ShortenRequest,
  UpdateCampaignRequest,
} from "@/api/types"

export const CAMPAIGNS_KEY = ["campaigns"] as const
export const PAGE_SIZE = 10

/** null status means "any status", matching the dropdown's default. */
export type CampaignFilters = {
  search: string
  status: CampaignStatus | null
}

export function useCampaigns(page: number, filters: CampaignFilters) {
  return useQuery({
    // The filters belong in the key, or a filtered request would be answered
    // from the unfiltered cache entry.
    queryKey: [...CAMPAIGNS_KEY, "list", page, filters.search, filters.status],
    queryFn: () =>
      campaignsApi.list({
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
        search: filters.search || undefined,
        status: filters.status ?? undefined,
      }),
    // Without this the table drops to its skeleton on every debounce tick.
    placeholderData: keepPreviousData,
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
