import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { SearchEngine, SearchEngineCreate, SearchEngineUpdate } from '@/types/search-engine'

export const searchEngineKeys = {
  all: ['search-engines'] as const,
}

export function useSearchEngines() {
  return useQuery({
    queryKey: searchEngineKeys.all,
    queryFn: () => apiClient.get<SearchEngine[]>('api/search-engines'),
  })
}

export function useCreateSearchEngine() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: SearchEngineCreate) => apiClient.post<SearchEngine>('api/search-engines', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: searchEngineKeys.all })
    },
  })
}

export function useUpdateSearchEngine() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: SearchEngineUpdate }) =>
      apiClient.put<SearchEngine>(`api/search-engines/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: searchEngineKeys.all })
    },
  })
}

export function useDeleteSearchEngine() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`api/search-engines/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: searchEngineKeys.all })
    },
  })
}
