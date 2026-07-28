export type ExternalRequestMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
export type ExternalRequestBodyType = 'none' | 'json' | 'form' | 'text' | 'raw'
export type ExternalRequestParserType = 'status' | 'text' | 'json_path'

export interface ExternalRequest {
  id: string
  name: string
  description?: string
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
  order_index: number
  created_at?: string
  updated_at?: string
}

export interface ExternalRequestCreate {
  name: string
  description?: string
  method: ExternalRequestMethod
  url: string
  headers_json?: string
  query_json?: string
  body_type?: ExternalRequestBodyType
  body?: string
  parser_type?: ExternalRequestParserType
  parser_config_json?: string
  confirm_before_run?: boolean
  enabled?: boolean
  order_index?: number
}

export type ExternalRequestUpdate = Partial<ExternalRequestCreate>

export interface ExternalRequestParsedField {
  label: string
  path?: string
  value?: unknown
  error?: string
}

export interface ExternalRequestExecuteResult {
  status: number
  status_text: string
  duration_ms: number
  headers: Record<string, string[]>
  body_preview: string
  parsed: ExternalRequestParsedField[]
  error?: string
}
