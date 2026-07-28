import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { Link, LinkFlow, LinkFlowCreate, LinkFlowUpdate } from '@/types/link'
import { linkKeys } from './use-links'

export const linkFlowKeys = {
  all: ['link-flows'] as const,
}

export function useCreateLinkFlow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: LinkFlowCreate) => apiClient.post<LinkFlow>('api/link-flows', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useUpdateLinkFlow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: LinkFlowUpdate }) =>
      apiClient.put<LinkFlow>(`api/link-flows/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useDeleteLinkFlow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id }: { id: string; linkIdsToKeep?: string[] }) =>
      apiClient.delete(`api/link-flows/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useUpdateFlowLinkOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ flowId, linkId, orderIndex }: { flowId: string; linkId: string; orderIndex: number }) =>
      apiClient.put<Link>(`api/link-flows/${flowId}/links/${linkId}`, { order_index: orderIndex }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useAddLinkToFlow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ flowId, linkId, orderIndex }: { flowId: string; linkId: string; orderIndex?: number }) =>
      apiClient.post<Link>(`api/link-flows/${flowId}/links`, { link_id: linkId, order_index: orderIndex }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}

export function useRemoveLinkFromFlow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ flowId, linkId }: { flowId: string; linkId: string }) =>
      apiClient.delete(`api/link-flows/${flowId}/links/${linkId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.byGroup })
    },
  })
}
