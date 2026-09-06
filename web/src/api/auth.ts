import { API_URL, client } from "./client"
import type { User } from "./types"

export const authApi = {
  /**
   * OAuth needs a full page navigation - the consent screen can't be
   * completed from an XHR.
   */
  loginUrl: () => `${API_URL}/api/auth/google/login`,

  async me(): Promise<User> {
    const { data } = await client.get<{ user: User }>("/api/auth/me")
    return data.user
  },

  async logout(): Promise<void> {
    await client.post("/api/auth/logout")
  },
}
