import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { dataApi, type ImportMode, type ImportResult } from '@/lib/api-client'
import { Download, RotateCcw, Upload } from 'lucide-react'
import { toast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

export default function DataManagement() {
  const [isExporting, setIsExporting] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importMode, setImportMode] = useState<ImportMode>('merge')
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const { t } = useTranslation()

  const handleExport = async () => {
    try {
      setIsExporting(true)
      const blob = await dataApi.exportAll()
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `hometab_backup_${new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)}.zip`
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Export failed:', error)
      toast({ title: t('data.exportFailed'), description: t('data.exportFailedDesc'), variant: 'destructive' })
    } finally {
      setIsExporting(false)
    }
  }

  const handleImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    if (importMode === 'replace' && !window.confirm(t('data.replaceConfirm'))) {
      event.target.value = ''
      return
    }

    try {
      setIsImporting(true)
      setImportResult(null)
      const result = await dataApi.importAll(file, importMode)
      setImportResult(result)
      const imported = Object.values(result.imported).reduce((sum, count) => sum + count, 0)
      const skipped = Object.values(result.skipped).reduce((sum, count) => sum + count, 0)
      toast({
        title: t('data.importSuccess'),
        description: t('data.importSuccessDesc', { imported, skipped }),
      })
    } catch (error) {
      console.error('Import failed:', error)
      toast({
        title: t('data.importFailed'),
        description: error instanceof Error ? error.message : t('data.importFailedDesc'),
        variant: 'destructive',
      })
    } finally {
      setIsImporting(false)
      event.target.value = ''
    }
  }

  return (
    <div className="space-y-4">
      <RadioGroup value={importMode} onValueChange={(value) => setImportMode(value as ImportMode)} className="grid gap-2 sm:grid-cols-2">
        <Label htmlFor="import-merge" className="flex cursor-pointer items-start gap-3 rounded-md border p-3">
          <RadioGroupItem id="import-merge" value="merge" className="mt-0.5" />
          <span>
            <span className="block text-sm font-medium">{t('data.mergeMode')}</span>
            <span className="block text-xs text-muted-foreground">{t('data.mergeModeDesc')}</span>
          </span>
        </Label>
        <Label htmlFor="import-replace" className="flex cursor-pointer items-start gap-3 rounded-md border border-amber-300 p-3">
          <RadioGroupItem id="import-replace" value="replace" className="mt-0.5" />
          <span>
            <span className="block text-sm font-medium">{t('data.replaceMode')}</span>
            <span className="block text-xs text-muted-foreground">{t('data.replaceModeDesc')}</span>
          </span>
        </Label>
      </RadioGroup>

      <div className="flex flex-wrap gap-2">
        <Label htmlFor="import-file" className="sr-only">{t('data.selectFile')}</Label>
        <Input
          id="import-file"
          type="file"
          accept=".zip,application/zip"
          onChange={handleImport}
          disabled={isImporting}
          className="hidden"
        />
        <Button
          variant={importMode === 'replace' ? 'destructive' : 'outline'}
          size="sm"
          onClick={() => document.getElementById('import-file')?.click()}
          disabled={isImporting}
        >
          {importMode === 'replace' ? <RotateCcw className="h-4 w-4" /> : <Upload className="h-4 w-4" />}
          {isImporting ? t('data.importing') : t('data.unifiedImport')}
        </Button>

        <Button variant="outline" size="sm" onClick={handleExport} disabled={isExporting}>
          <Download className="h-4 w-4" />
          {isExporting ? t('data.exporting') : t('data.unifiedExport')}
        </Button>
      </div>

      {importResult && (
        <div className="space-y-1 rounded-md border bg-muted/40 p-3 text-xs text-muted-foreground">
          <div>{t('data.importCount', { count: Object.values(importResult.imported).reduce((sum, n) => sum + n, 0) })}</div>
          <div>{t('data.skipCount', { count: Object.values(importResult.skipped).reduce((sum, n) => sum + n, 0) })}</div>
          {importResult.errors.length > 0 && <div className="text-destructive">{t('data.errorCount', { count: importResult.errors.length })}</div>}
          {importResult.pre_restore_backup && (
            <a
              href={dataApi.backupDownloadUrl(importResult.pre_restore_backup.id)}
              className="inline-flex items-center gap-1 font-medium text-primary underline"
            >
              <Download className="h-3 w-3" />
              {t('data.downloadPreRestore')}
            </a>
          )}
        </div>
      )}
    </div>
  )
}
