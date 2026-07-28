import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { LinkGroup, LinkGroupCreate, LinkGroupUpdate } from '@/types/link'
import { linkKeys } from './use-links'

export const linkGroupKeys = {
  all: ['link-groups'] as const,
}

export function useLinkGroups() {
  return useQuery({
    queryKey: linkGroupKeys.all,
    queryFn: () => apiClient.get<LinkGroup[]>('api/link-groups'),
  })
}

export function useCreateLinkGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: LinkGroupCreate) => apiClient.post<LinkGroup>('api/link-groups', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkGroupKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useUpdateLinkGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: LinkGroupUpdate }) =>
      apiClient.put<LinkGroup>(`api/link-groups/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkGroupKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useDeleteLinkGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`api/link-groups/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkGroupKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}
