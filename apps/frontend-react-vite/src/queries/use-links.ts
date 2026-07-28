import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { Link, LinkCreate, LinkUpdate, GroupedLinks } from '@/types/link'

export const linkKeys = {
  all: ['links'] as const,
  byGroup: ['links-by-group'] as const,
}

export function useLinks() {
  return useQuery({
    queryKey: linkKeys.all,
    queryFn: () => apiClient.get<Link[]>('api/links'),
  })
}

export function useLinksByGroup() {
  return useQuery({
    queryKey: linkKeys.byGroup,
    queryFn: () => apiClient.get<GroupedLinks[]>('api/links-by-group'),
  })
}

export function useCreateLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: LinkCreate) => apiClient.post<Link>('api/links', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useUpdateLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: LinkUpdate }) =>
      apiClient.put<Link>(`api/links/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useDeleteLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`api/links/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.all })
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}
