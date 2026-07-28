import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { SettingRead, SettingCreate } from '@/types/setting'

export const settingKeys = {
  all: ['settings'] as const,
  byKey: (key: string) => ['settings', key] as const,
}

export function useSettings() {
  return useQuery({
    queryKey: settingKeys.all,
    queryFn: () => apiClient.get<Record<string, unknown>>('api/settings'),
  })
}

export function useSetting(key: string) {
  return useQuery({
    queryKey: settingKeys.byKey(key),
    queryFn: () => apiClient.get<SettingRead>(`api/settings/${key}`),
    enabled: !!key,
  })
}

export function useUpdateSetting() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: SettingCreate) => apiClient.post<SettingRead>('api/settings', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: settingKeys.all })
    },
  })
}

export function useDeleteSetting() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (key: string) => apiClient.delete(`api/settings/${key}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: settingKeys.all })
    },
  })
}
