import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { authApi } from "@/api/auth"
import { ApiError } from "@/api/client"

export const authQueryKey = ["auth", "me"] as const

/**
 * A 401 here is a normal state ("signed out"), not a failure, so it must not
 * be retried or surfaced as an error.
 */
export function useAuth() {
  const query = useQuery({
    queryKey: authQueryKey,
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  const isUnauthenticated = query.error instanceof ApiError && query.error.isUnauthorized

  return {
    user: query.data ?? null,
    isLoading: query.isLoading,
    isAuthenticated: Boolean(query.data),
    isUnauthenticated,
    error: isUnauthenticated ? null : query.error,
  }
}

export function useLogout() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      queryClient.clear()
    },
  })
}
