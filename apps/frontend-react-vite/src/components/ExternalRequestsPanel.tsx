import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Edit2,
  FileJson,
  Loader2,
  Plus,
} from 'lucide-react'
import { Button } from './ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog'
import { Label } from './ui/label'
import {
  useCreateExternalRequest,
  useDeleteExternalRequest,
  useExecuteExternalRequest,
  useExternalRequests,
  useUpdateExternalRequest,
} from '@/queries/use-external-requests'
import type {
  ExternalRequest,
  ExternalRequestCreate,
  ExternalRequestExecuteResult,
} from '@/types/external-request'
import { useTranslation } from 'react-i18next'
import { RequestDialog } from './external-requests/RequestDialog'

function formatValue(value: unknown, t: (key: string) => string) {
  if (value === null || typeof value === 'undefined') return t('external.emptyValue')
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

function getResultTone(result?: ExternalRequestExecuteResult) {
  if (!result) return 'text-gray-500'
  if (result.error || result.status >= 400 || result.status === 0) return 'text-red-600'
  return 'text-emerald-700'
}

export default function ExternalRequestsPanel() {
  const { data: requests = [], isLoading } = useExternalRequests()
  const createRequest = useCreateExternalRequest()
  const updateRequest = useUpdateExternalRequest()
  const deleteRequest = useDeleteExternalRequest()
  const executeRequest = useExecuteExternalRequest()
  const { t } = useTranslation()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRequest, setEditingRequest] = useState<ExternalRequest | null>(null)
  const [pendingRun, setPendingRun] = useState<ExternalRequest | null>(null)
  const [pendingDelete, setPendingDelete] = useState<ExternalRequest | null>(null)
  const [detailResult, setDetailResult] = useState<{ request: ExternalRequest; result: ExternalRequestExecuteResult } | null>(null)
  const [results, setResults] = useState<Record<string, ExternalRequestExecuteResult>>({})

  const nextOrderIndex = useMemo(() => {
    if (requests.length === 0) return 0
    return Math.max(...requests.map((request) => request.order_index || 0)) + 10
  }, [requests])

  const sortedRequests = useMemo(
    () => [...requests].sort((a, b) => a.order_index - b.order_index || a.name.localeCompare(b.name)),
    [requests],
  )

  const openCreateDialog = () => {
    setEditingRequest(null)
    setDialogOpen(true)
  }

  const handleSave = async (payload: ExternalRequestCreate) => {
    if (editingRequest) {
      await updateRequest.mutateAsync({ id: editingRequest.id, data: payload })
      return
    }
    await createRequest.mutateAsync({ ...payload, order_index: nextOrderIndex })
  }

  const runRequest = async (request: ExternalRequest) => {
    const result = await executeRequest.mutateAsync(request.id)
    setResults((current) => ({ ...current, [request.id]: result }))
    setDetailResult({ request, result })
  }

  const handleRun = (request: ExternalRequest) => {
    if (!request.enabled) return
    if (request.confirm_before_run) {
      setPendingRun(request)
      return
    }
    runRequest(request).catch((err) => {
      const result: ExternalRequestExecuteResult = {
        status: 0,
        status_text: t('external.executeFailed'),
        duration_ms: 0,
        headers: {},
        body_preview: '',
        parsed: [{ label: t('external.errorLabel'), error: err instanceof Error ? err.message : t('external.executeFailed') }],
        error: err instanceof Error ? err.message : t('external.executeFailed'),
      }
      setResults((current) => ({ ...current, [request.id]: result }))
    })
  }

  const handleCopy = async (request: ExternalRequest) => {
    await createRequest.mutateAsync({
      name: `${request.name} ${t('external.copySuffix')}`,
      description: request.description,
      method: request.method,
      url: request.url,
      headers_json: request.headers_json,
      query_json: request.query_json,
      body_type: request.body_type,
      body: request.body,
      parser_type: request.parser_type,
      parser_config_json: request.parser_config_json,
      confirm_before_run: request.confirm_before_run,
      enabled: request.enabled,
      order_index: nextOrderIndex,
    })
  }

  const handleConfirmRun = async () => {
    if (!pendingRun) return
    const target = pendingRun
    setPendingRun(null)
    await runRequest(target)
  }

  const handleConfirmDelete = async () => {
    if (!pendingDelete) return
    const target = pendingDelete
    setPendingDelete(null)
    await deleteRequest.mutateAsync(target.id)
    setResults((current) => {
      const next = { ...current }
      delete next[target.id]
      return next
    })
  }

  return (
    <section className="mb-5 rounded-2xl border border-white/30 bg-white/20 p-4 shadow-lg backdrop-blur-sm">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <h2 className="text-base font-medium text-white/90 drop-shadow-sm">{t('external.requestTitle')}</h2>
          <span className="text-xs text-white/55 bg-white/10 px-1.5 py-0.5 rounded-full">{sortedRequests.length}</span>
        </div>
        <Button
          type="button"
          variant="outline"
          onClick={openCreateDialog}
          className="border-white/30 bg-white/20 text-white backdrop-blur-sm hover:bg-white/30 hover:text-white"
        >
          <Plus className="mr-2 h-4 w-4" />{t('external.addRequest')}
        </Button>
      </div>

      {isLoading ? (
        <div className="rounded-xl border border-white/20 bg-white/10 px-4 py-6 text-center text-sm text-white/60">
          {t('external.loading')}
        </div>
      ) : sortedRequests.length === 0 ? (
        <div className="rounded-xl border border-dashed border-white/30 bg-white/10 px-4 py-7 text-center">
          <FileJson className="mx-auto mb-2 h-7 w-7 text-white/45" />
          <p className="text-sm text-white/70">{t('external.noRequests')}</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-3">
          {sortedRequests.map((request) => {
            const result = results[request.id]
            const running = executeRequest.isPending && executeRequest.variables === request.id
            return (
              <div
                key={request.id}
                role="button"
                tabIndex={request.enabled && !running ? 0 : -1}
                aria-disabled={!request.enabled || running}
                onClick={() => {
                  if (!request.enabled || running) return
                  handleRun(request)
                }}
                onKeyDown={(event) => {
                  if (!request.enabled || running) return
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    handleRun(request)
                  }
                }}
                className={`group relative min-h-10 overflow-hidden rounded-xl border border-white/30 px-3 py-2.5 text-left backdrop-blur-sm transition-all duration-200 ${
                  request.enabled
                    ? 'bg-white/90 text-gray-900 hover:bg-white hover:shadow-md cursor-pointer'
                    : 'bg-white/45 text-gray-500 cursor-not-allowed'
                }`}
                title={request.url}
              >
                <div className="flex items-center gap-2 pr-9">
                  <span className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-sm ${
                    result ? getResultTone(result) : 'text-cyan-700'
                  } bg-cyan-50`}>
                    {running ? <Loader2 className="h-3 w-3 animate-spin" /> : <Activity className="h-3 w-3" />}
                  </span>
                  <span className="truncate text-sm font-medium">{request.name}</span>
                </div>
                <div className="absolute right-0 top-0 bottom-0 w-16 bg-gradient-to-l from-white via-white/95 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                <div className="absolute right-2 top-1/2 transform -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                  <button
                    type="button"
                    title={t('external.editRequest')}
                    className="inline-flex h-6 w-6 items-center justify-center rounded-md bg-white/90 text-gray-600 shadow-sm transition-colors hover:bg-white hover:text-gray-800"
                    onClick={(event) => {
                      event.stopPropagation()
                      setEditingRequest(request)
                      setDialogOpen(true)
                    }}
                  >
                    <Edit2 className="h-3 w-3" />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      <RequestDialog
        open={dialogOpen}
        request={editingRequest}
        onOpenChange={setDialogOpen}
        onSave={handleSave}
        onCopy={handleCopy}
        onDelete={setPendingDelete}
      />

      <Dialog open={Boolean(pendingRun)} onOpenChange={(open) => !open && setPendingRun(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
              {t('external.executeRequest')}
            </DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {t('external.confirmExecute', { method: pendingRun?.method, name: pendingRun?.name })}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingRun(null)}>{t('external.cancel')}</Button>
            <Button onClick={handleConfirmRun}>{t('external.execute')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(pendingDelete)} onOpenChange={(open) => !open && setPendingDelete(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              {t('external.deleteRequest')}
            </DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {t('external.confirmDelete', { name: pendingDelete?.name })}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingDelete(null)}>{t('external.cancel')}</Button>
            <Button variant="destructive" onClick={handleConfirmDelete}>{t('external.delete')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(detailResult)} onOpenChange={(open) => !open && setDetailResult(null)}>
        <DialogContent className="w-[calc(100vw-2rem)] sm:max-w-2xl max-h-[88vh] overflow-y-auto overflow-x-hidden">
          <DialogHeader>
            <DialogTitle>{detailResult?.request.name || t('external.resultTitle')}</DialogTitle>
          </DialogHeader>
          {detailResult && (
            <div className="space-y-4 text-sm min-w-0">
              <div className="grid gap-2 sm:grid-cols-3">
                <div className="min-w-0 rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">{t('external.status')}</div>
                  <div className="font-medium break-words">{detailResult.result.status_text || detailResult.result.status || 'ERR'}</div>
                </div>
                <div className="min-w-0 rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">{t('external.duration')}</div>
                  <div className="font-medium">{detailResult.result.duration_ms}ms</div>
                </div>
                <div className="min-w-0 rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">{t('external.parsed')}</div>
                  <div className="font-medium">{t('external.items', { count: detailResult.result.parsed.length })}</div>
                </div>
              </div>

              <div className="space-y-2 min-w-0">
                <Label>{t('external.parseResult')}</Label>
                <div className="min-w-0 rounded-lg border bg-muted/30 p-3">
                  {detailResult.result.parsed.length > 0 ? (
                    <div className="space-y-2">
                      {detailResult.result.parsed.map((field) => (
                        <div key={`${field.label}-${field.path || 'value'}`} className="flex min-w-0 gap-2">
                          <span className="w-28 shrink-0 text-muted-foreground">{field.label}</span>
                          <span className="min-w-0 break-words [overflow-wrap:anywhere]">{field.error || formatValue(field.value, t)}</span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <span className="text-muted-foreground">{t('external.none')}</span>
                  )}
                </div>
              </div>

              <div className="space-y-2 min-w-0">
                <Label>{t('external.responseSummary')}</Label>
                <pre className="max-h-64 max-w-full overflow-y-auto overflow-x-hidden whitespace-pre-wrap break-words [overflow-wrap:anywhere] rounded-lg border bg-slate-950 p-3 text-xs text-slate-100">
                  {detailResult.result.body_preview || detailResult.result.error || t('external.noResponseBody')}
                </pre>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </section>
  )
}
