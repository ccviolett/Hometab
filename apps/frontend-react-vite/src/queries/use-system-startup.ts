import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { StartupConfig, StartupStatus } from '@/types/system'

export const systemStartupKey = ['system', 'startup'] as const

export function useSystemStartup(enabled = true) {
  return useQuery({
    queryKey: systemStartupKey,
    queryFn: () => apiClient.get<StartupStatus>('api/system/startup'),
    enabled,
    staleTime: 0,
  })
}

export function useConfigureSystemStartup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (config: StartupConfig) =>
      apiClient.put<StartupStatus>('api/system/startup', config),
    onSuccess: (status) => {
      queryClient.setQueryData(systemStartupKey, status)
    },
  })
}
