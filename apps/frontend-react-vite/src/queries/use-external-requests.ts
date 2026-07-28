import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type {
  ExternalRequest,
  ExternalRequestCreate,
  ExternalRequestExecuteResult,
  ExternalRequestUpdate,
} from '@/types/external-request'

export const externalRequestKeys = {
  all: ['external-requests'] as const,
}

export function useExternalRequests() {
  return useQuery({
    queryKey: externalRequestKeys.all,
    queryFn: () => apiClient.get<ExternalRequest[]>('api/external-requests'),
  })
}

export function useCreateExternalRequest() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: ExternalRequestCreate) => apiClient.post<ExternalRequest>('api/external-requests', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: externalRequestKeys.all })
    },
  })
}

export function useUpdateExternalRequest() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ExternalRequestUpdate }) =>
      apiClient.put<ExternalRequest>(`api/external-requests/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: externalRequestKeys.all })
    },
  })
}

export function useDeleteExternalRequest() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`api/external-requests/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: externalRequestKeys.all })
    },
  })
}

export function useExecuteExternalRequest() {
  return useMutation({
    mutationFn: (id: string) => apiClient.post<ExternalRequestExecuteResult>(`api/external-requests/${id}/execute`),
  })
}
