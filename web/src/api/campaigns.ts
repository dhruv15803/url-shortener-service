import { client } from "./client"
import type {
  AddCampaignRequest,
  Campaign,
  CampaignListResponse,
  DestinationRef,
  ShortUrlRef,
  ShortenRequest,
  ShortenResult,
  UpdateCampaignRequest,
} from "./types"

type ShortenCreatedBody = { destination_url: DestinationRef; short_url: ShortUrlRef }
type ShortenExistingBody = { destination_url: DestinationRef; short_urls: ShortUrlRef[] }

export const campaignsApi = {
  async list(params: { limit: number; offset: number }): Promise<CampaignListResponse> {
    const { data } = await client.get<CampaignListResponse>("/api/urls", { params })
    return data
  },

  async get(shortCode: string): Promise<Campaign> {
    const { data } = await client.get<{ campaign: Campaign }>(`/api/urls/${shortCode}`)
    return data.campaign
  },

  async update(shortCode: string, patch: UpdateCampaignRequest): Promise<Campaign> {
    const { data } = await client.put<{ campaign: Campaign }>(`/api/urls/${shortCode}`, patch)
    return data.campaign
  },

  /**
   * 201 means a new destination was created with its first campaign; 200 means
   * the user already had this destination, and its existing campaigns come
   * back instead. The status code is the only reliable discriminant, so it is
   * resolved here rather than leaking into components.
   */
  async shorten(body: ShortenRequest): Promise<ShortenResult> {
    const response = await client.post<ShortenCreatedBody | ShortenExistingBody>("/api/urls", body)

    if (response.status === 201) {
      const created = response.data as ShortenCreatedBody
      return { created: true, destination: created.destination_url, shortUrl: created.short_url }
    }

    const existing = response.data as ShortenExistingBody
    return { created: false, destination: existing.destination_url, shortUrls: existing.short_urls }
  },

  async addToDestination(destinationId: number, body: AddCampaignRequest): Promise<ShortUrlRef> {
    const { data } = await client.post<{ short_url: ShortUrlRef }>(
      `/api/destinations/${destinationId}/short-urls`,
      body,
    )
    return data.short_url
  },
}
