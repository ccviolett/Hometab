export interface Link {
  id: string
  name: string
  url: string
  group_id?: string
  flow_id?: string | null
  order_index: number
  created_at?: string
  updated_at?: string
}

export interface LinkCreate {
  name: string
  url: string
  group_id?: string
  flow_id?: string | null
  order_index?: number
}

export interface LinkUpdate {
  name?: string
  url?: string
  group_id?: string
  flow_id?: string | null
  order_index?: number
}

export interface LinkGroup {
  id: string
  name: string
  description?: string
  order_index: number
  created_at?: string
  updated_at?: string
}

export interface LinkGroupCreate {
  name: string
  description?: string
  order_index?: number
}

export interface LinkGroupUpdate {
  name?: string
  description?: string
  order_index?: number
}

export interface LinkFlow {
  id: string
  group_id: string
  name: string
  order_index: number
  created_at?: string
  updated_at?: string
}

export interface LinkFlowCreate {
  group_id: string
  name: string
  order_index?: number
}

export interface LinkFlowUpdate {
  group_id?: string
  name?: string
  order_index?: number
}

export interface LinkFlowWithLinks {
  flow: LinkFlow
  links: Link[]
}

export interface GroupedLinks {
  group: LinkGroup
  flows: LinkFlowWithLinks[]
  links: Link[]
}
