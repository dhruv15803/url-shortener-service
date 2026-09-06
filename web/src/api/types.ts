import type { components } from "./schema"

/**
 * Readable aliases over the types generated from ../../openapi.yaml.
 * Regenerate with: npm run gen:api
 */
type Schemas = components["schemas"]

export type User = Schemas["User"]
export type Campaign = Schemas["Campaign"]
export type DestinationRef = Schemas["DestinationRef"]
export type ShortUrlRef = Schemas["ShortURLRef"]
export type CampaignStatus = Schemas["ShortURLStatus"]

export type CampaignListResponse = Schemas["CampaignListResponse"]
export type ShortenRequest = Schemas["CreateDestinationRequest"]
export type AddCampaignRequest = Schemas["CreateShortURLRequest"]
export type UpdateCampaignRequest = Schemas["UpdateCampaignRequest"]

/**
 * POST /api/urls answers with a different shape depending on whether the
 * destination already existed, so the status code is modelled as a
 * discriminant rather than left for callers to sniff.
 */
export type ShortenResult =
  | { created: true; destination: DestinationRef; shortUrl: ShortUrlRef }
  | { created: false; destination: DestinationRef; shortUrls: ShortUrlRef[] }

/** The two values PUT accepts. The other statuses are derived, never sent. */
export type StatusIntent = "ACTIVE" | "DISABLED"
