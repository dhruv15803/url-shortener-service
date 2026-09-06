import axios, { AxiosError } from "axios"

export const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080"

/**
 * withCredentials is required: the session lives in an httpOnly cookie, and
 * the api is on a different origin than the dev server.
 */
export const client = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
})

/** An api failure with the server's `{ error }` message already unwrapped. */
export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }

  get isUnauthorized() {
    return this.status === 401
  }
}

client.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string }>) => {
    if (error.response) {
      const message = error.response.data?.error ?? error.message
      return Promise.reject(new ApiError(error.response.status, message))
    }

    // No response at all - network failure, server down, or a blocked
    // cross-origin request.
    return Promise.reject(new ApiError(0, "Could not reach the server"))
  },
)
