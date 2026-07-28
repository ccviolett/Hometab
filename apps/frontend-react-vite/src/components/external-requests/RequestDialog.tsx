import { useEffect, useState } from 'react'
import { Copy, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Switch } from '../ui/switch'
import { RowsEditor } from './RowsEditor'
import {
  keyValueRowsFromJSON,
  keyValueRowsToJSON,
  parserRowsFromJSON,
  parserRowsToJSON,
  type KeyValueRow,
  type ParserRow,
} from './request-form-utils'
import type {
  ExternalRequest,
  ExternalRequestBodyType,
  ExternalRequestCreate,
  ExternalRequestMethod,
  ExternalRequestParserType,
} from '@/types/external-request'

const methods: ExternalRequestMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
const bodyTypes: ExternalRequestBodyType[] = ['none', 'json', 'form', 'text', 'raw']
const parserTypes: ExternalRequestParserType[] = ['status', 'text', 'json_path']

interface RequestFormState {
  name: string
  description: string
  method: ExternalRequestMethod
  url: string
  headers_json: string
  query_json: string
  body_type: ExternalRequestBodyType
  body: string
  parser_type: ExternalRequestParserType
  parser_config_json: string
  confirm_before_run: boolean
  enabled: boolean
}

const emptyForm: RequestFormState = {
  name: '',
  description: '',
  method: 'GET',
  url: '',
  headers_json: '',
  query_json: '',
  body_type: 'none',
  body: '',
  parser_type: 'status',
  parser_config_json: '',
  confirm_before_run: false,
  enabled: true,
}

function formFromRequest(request: ExternalRequest | null): RequestFormState {
  if (!request) return emptyForm
  return {
    name: request.name,
    description: request.description || '',
    method: request.method,
    url: request.url,
    headers_json: request.headers_json || '',
    query_json: request.query_json || '',
    body_type: request.body_type || 'none',
    body: request.body || '',
    parser_type: request.parser_type || 'status',
    parser_config_json: request.parser_config_json || '',
    confirm_before_run: request.confirm_before_run,
    enabled: request.enabled,
  }
}

function toPayload(form: RequestFormState, orderIndex?: number): ExternalRequestCreate {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    method: form.method,
    url: form.url.trim(),
    headers_json: form.headers_json.trim(),
    query_json: form.query_json.trim(),
    body_type: form.body_type,
    body: form.body,
    parser_type: form.parser_type,
    parser_config_json: form.parser_config_json.trim(),
    confirm_before_run: form.confirm_before_run,
    enabled: form.enabled,
    order_index: orderIndex,
  }
}

function textAreaClassName() {
  return 'min-h-[74px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2'
}

export function RequestDialog({
  open,
  request,
  onOpenChange,
  onSave,
  onCopy,
  onDelete,
}: {
  open: boolean
  request: ExternalRequest | null
  onOpenChange: (open: boolean) => void
  onSave: (payload: ExternalRequestCreate) => Promise<void>
  onCopy: (request: ExternalRequest) => Promise<void>
  onDelete: (request: ExternalRequest) => void
}) {
  const [form, setForm] = useState<RequestFormState>(emptyForm)
  const [headerRows, setHeaderRows] = useState<KeyValueRow[]>(keyValueRowsFromJSON(''))
  const [queryRows, setQueryRows] = useState<KeyValueRow[]>(keyValueRowsFromJSON(''))
  const [parserRows, setParserRows] = useState<ParserRow[]>(parserRowsFromJSON(''))
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const { t } = useTranslation()

  useEffect(() => {
    if (open) {
      const nextForm = formFromRequest(request)
      setForm(nextForm)
      setHeaderRows(keyValueRowsFromJSON(nextForm.headers_json))
      setQueryRows(keyValueRowsFromJSON(nextForm.query_json))
      setParserRows(parserRowsFromJSON(nextForm.parser_config_json))
      setError('')
    }
  }, [open, request])

  const update = <K extends keyof RequestFormState>(key: K, value: RequestFormState[K]) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const handleMethodChange = (method: ExternalRequestMethod) => {
    setForm((current) => ({
      ...current,
      method,
      confirm_before_run: request ? current.confirm_before_run : method !== 'GET',
    }))
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!form.name.trim() || !form.url.trim()) {
      setError(t('external.nameUrlRequired'))
      return
    }
    setSaving(true)
    setError('')
    try {
      const normalizedForm = {
        ...form,
        headers_json: keyValueRowsToJSON(headerRows),
        query_json: keyValueRowsToJSON(queryRows),
        parser_config_json: form.parser_type === 'json_path' ? parserRowsToJSON(parserRows) : '',
      }
      if (normalizedForm.body_type === 'json' && normalizedForm.body.trim()) JSON.parse(normalizedForm.body)
      await onSave(toPayload(normalizedForm))
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('external.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[88vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{request ? t('external.editRequest') : t('external.addRequest')}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-[120px_1fr]">
            <div className="space-y-2">
              <Label htmlFor="request-method">{t('external.method')}</Label>
              <select
                id="request-method"
                value={form.method}
                onChange={(event) => handleMethodChange(event.target.value as ExternalRequestMethod)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {methods.map((method) => (
                  <option key={method} value={method}>{method}</option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="request-url">URL</Label>
              <Input
                id="request-url"
                value={form.url}
                onChange={(event) => update('url', event.target.value)}
                placeholder="http://127.0.0.1:3999/health"
                required
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="request-name">{t('external.name')}</Label>
              <Input
                id="request-name"
                value={form.name}
                onChange={(event) => update('name', event.target.value)}
                placeholder={t('external.namePlaceholder')}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="request-description">{t('external.description')}</Label>
              <Input
                id="request-description"
                value={form.description}
                onChange={(event) => update('description', event.target.value)}
                placeholder={t('external.descriptionPlaceholder')}
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('external.headers')}</Label>
              <RowsEditor
                rows={headerRows}
                columns={[{ key: 'key', placeholder: t('external.key') }, { key: 'value', placeholder: t('external.value') }]}
                onChange={setHeaderRows}
                addLabel={t('external.addRow')}
                removeLabel={t('external.removeRow')}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('external.query')}</Label>
              <RowsEditor
                rows={queryRows}
                columns={[{ key: 'key', placeholder: t('external.key') }, { key: 'value', placeholder: t('external.value') }]}
                onChange={setQueryRows}
                addLabel={t('external.addRow')}
                removeLabel={t('external.removeRow')}
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-[140px_1fr]">
            <div className="space-y-2">
              <Label htmlFor="request-body-type">{t('external.bodyType')}</Label>
              <select
                id="request-body-type"
                value={form.body_type}
                onChange={(event) => update('body_type', event.target.value as ExternalRequestBodyType)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {bodyTypes.map((type) => (
                  <option key={type} value={type}>{type}</option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="request-body">{t('external.body')}</Label>
              <textarea
                id="request-body"
                value={form.body}
                onChange={(event) => update('body', event.target.value)}
                className={textAreaClassName()}
                placeholder='{"action":"clear"}'
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-[140px_1fr]">
            <div className="space-y-2">
              <Label htmlFor="request-parser">{t('external.parser')}</Label>
              <select
                id="request-parser"
                value={form.parser_type}
                onChange={(event) => update('parser_type', event.target.value as ExternalRequestParserType)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {parserTypes.map((type) => (
                  <option key={type} value={type}>{type}</option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label>{t('external.parserConfig')}</Label>
              {form.parser_type === 'json_path' ? (
                <RowsEditor
                  rows={parserRows}
                  columns={[{ key: 'label', placeholder: t('external.parserLabel') }, { key: 'path', placeholder: '$.data.total' }]}
                  onChange={setParserRows}
                  addLabel={t('external.addRow')}
                  removeLabel={t('external.removeRow')}
                />
              ) : (
                <p className="py-2 text-sm text-muted-foreground">{t('external.parserConfigNotRequired')}</p>
              )}
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-5 rounded-lg border bg-muted/30 px-3 py-3">
            <label className="flex items-center gap-2 text-sm">
              <Switch checked={form.confirm_before_run} onCheckedChange={(checked) => update('confirm_before_run', checked)} />
              {t('external.confirmBeforeRun')}
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Switch checked={form.enabled} onCheckedChange={(checked) => update('enabled', checked)} />
              {t('external.enabled')}
            </label>
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <DialogFooter className="flex justify-between">
            <div className="flex gap-2">
              {request && (
                <>
                  <Button type="button" variant="destructive" size="sm" onClick={() => { onDelete(request); onOpenChange(false) }}>
                    <Trash2 className="h-4 w-4 mr-1" />{t('external.delete')}
                  </Button>
                  <Button type="button" variant="outline" size="sm" onClick={() => { onCopy(request); onOpenChange(false) }}>
                    <Copy className="h-4 w-4 mr-1" />{t('external.copy')}
                  </Button>
                </>
              )}
            </div>
            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>{t('external.cancel')}</Button>
              <Button type="submit" disabled={saving}>{saving ? t('external.saving') : t('external.save')}</Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
