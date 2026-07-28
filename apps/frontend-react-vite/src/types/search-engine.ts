export interface SearchEngine {
  id: number
  name: string
  url_template: string
  icon?: string
  description?: string
  color?: string
  created_at?: string
  updated_at?: string
}

export interface SearchEngineCreate {
  name: string
  url_template: string
  icon?: string
  description?: string
  color?: string
}

export interface SearchEngineUpdate {
  name?: string
  url_template?: string
  icon?: string
  description?: string
  color?: string
}
