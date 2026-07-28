import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { RefreshCw, Trash2, Upload } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import IconService, { type DomainIconItem } from '@/lib/iconService'
import { Button } from './ui/button'

export function IconManagementPanel({ refreshVersion = 0 }: { refreshVersion?: number }) {
  const { t } = useTranslation()
  const [items, setItems] = useState<DomainIconItem[]>([])
  const [filter, setFilter] = useState<'all' | DomainIconItem['status']>('all')
  const [loading, setLoading] = useState(true)
  const [busyHost, setBusyHost] = useState('')
  const [error, setError] = useState('')
  const uploadHost = useRef('')
  const fileInput = useRef<HTMLInputElement>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setItems(await IconService.list())
    } catch (err) {
      setError(err instanceof Error ? err.message : t('icons.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { void load() }, [load, refreshVersion])

  const filtered = useMemo(() => filter === 'all' ? items : items.filter((item) => item.status === filter), [items, filter])

  const run = async (host: string, action: () => Promise<unknown>) => {
    setBusyHost(host)
    setError('')
    try {
      await action()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('icons.actionFailed'))
    } finally {
      setBusyHost('')
    }
  }

  const handleFile = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    const host = uploadHost.current
    event.target.value = ''
    if (!file || !host) return
    await run(host, () => IconService.upload(host, file))
  }

  return (
    <div className="space-y-3 border-t pt-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-sm font-medium">{t('icons.title')}</h3>
        <div className="flex gap-1">
          {(['all', 'ready', 'failed', 'conflict'] as const).map((value) => (
            <Button key={value} type="button" size="sm" variant={filter === value ? 'default' : 'ghost'} onClick={() => setFilter(value)}>
              {t(`icons.filter.${value}`)}
            </Button>
          ))}
        </div>
      </div>
      <input ref={fileInput} type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml,image/x-icon,.ico" className="hidden" onChange={handleFile} />
      {loading ? <p className="text-sm text-muted-foreground">{t('external.loading')}</p> : filtered.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t('icons.empty')}</p>
      ) : (
        <div className="max-h-64 divide-y overflow-y-auto border-y">
          {filtered.map((item) => (
            <div key={item.host} className="flex min-h-14 items-center gap-3 py-2">
              <img src={`/api/link-icons/resolve?url=${encodeURIComponent(`https://${item.host}`)}&v=${refreshVersion}`} alt="" className="h-8 w-8 rounded object-contain" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-medium">{item.host}</div>
                <div className="text-xs text-muted-foreground">{item.source} · {item.status}</div>
                {item.error_message && <div className="truncate text-xs text-destructive" title={item.error_message}>{item.error_message}</div>}
              </div>
              <div className="flex gap-1">
                {item.status === 'failed' && (
                  <Button type="button" size="icon" variant="ghost" title={t('icons.retry')} disabled={busyHost === item.host} onClick={() => run(item.host, () => IconService.retry(item.host))}>
                    <RefreshCw className={busyHost === item.host ? 'animate-spin' : ''} />
                  </Button>
                )}
                {item.status === 'conflict' && (
                  <>
                    <Button type="button" size="sm" variant="outline" disabled={busyHost === item.host} onClick={() => run(item.host, () => IconService.chooseIcon(item.host, 'current'))}>{t('icons.keep')}</Button>
                    <Button type="button" size="sm" variant="outline" disabled={busyHost === item.host} onClick={() => run(item.host, () => IconService.chooseIcon(item.host, 'new'))}>{t('icons.useNew')}</Button>
                  </>
                )}
                <Button type="button" size="icon" variant="ghost" title={t('icons.upload')} disabled={busyHost === item.host} onClick={() => { uploadHost.current = item.host; fileInput.current?.click() }}>
                  <Upload />
                </Button>
                <Button type="button" size="icon" variant="ghost" title={t('icons.delete')} disabled={busyHost === item.host} onClick={() => run(item.host, () => IconService.remove(item.host))}>
                  <Trash2 />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
