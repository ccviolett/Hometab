import ky from 'ky'

export const api = ky.create({
  prefixUrl: '/',
  timeout: 30000,
  hooks: {
    beforeError: [
      async (error) => {
        const { response } = error
        if (response) {
          try {
            const body = await response.json() as Record<string, unknown>
            if (body.detail) {
              error.message = String(body.detail)
            }
          } catch {
            // ignore parse errors
          }
        }
        return error
      },
    ],
  },
})

export const apiClient = {
  get: <T>(url: string) =>
    api.get(url).json<T>(),

  post: <T>(url: string, json?: unknown) =>
    api.post(url, { json }).json<T>(),

  put: <T>(url: string, json?: unknown) =>
    api.put(url, { json }).json<T>(),

  delete: <T>(url: string) =>
    api.delete(url).json<T>(),
}

export type ImportMode = 'merge' | 'replace'

export interface ImportResult {
  imported: Record<string, number>
  skipped: Record<string, number>
  errors: Array<{ file: string; reason: string }>
  mode?: ImportMode
  format_version?: string
  pre_restore_backup?: { id: string; filename: string }
}

type ImportResponse = {
  result?: ImportResult
  imported?: Record<string, number>
  skipped?: Record<string, number>
  errors?: ImportResult['errors']
}

export const dataApi = {
  async exportAll(): Promise<Blob> {
    return api.get('api/export').blob()
  },
  async importAll(file: File, mode: ImportMode = 'merge'): Promise<ImportResult> {
    const formData = new FormData()
    formData.append('file', file)
    const data = await api.post(`api/import?mode=${mode}`, { body: formData }).json<ImportResponse>()
    return (
      data.result ?? {
        imported: data.imported ?? {},
        skipped: data.skipped ?? {},
        errors: data.errors ?? [],
      }
    )
  },
  backupDownloadUrl(id: string): string {
    return `/api/backups/${encodeURIComponent(id)}/download`
  },
}
