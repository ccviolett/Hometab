export interface Link {
  id: string;
  name: string;
  url: string;
  group_id?: string;
  flow_id?: string | null;
  order_index: number;
  created_at?: string;
  updated_at?: string;
}

export interface LinkGroup {
  id: string;
  name: string;
  description?: string;
  order_index: number;
  created_at?: string;
  updated_at?: string;
}

export interface LinkFlow {
  id: string;
  group_id: string;
  name: string;
  order_index: number;
  created_at?: string;
  updated_at?: string;
}

export interface LinkFlowWithLinks {
  flow: LinkFlow;
  links: Link[];
}

export interface GroupedLinks {
  group: LinkGroup;
  flows: LinkFlowWithLinks[];
  links: Link[];
}

export interface DragEndEvent {
  active: {
    id: string;
    data: {
      current?: {
        type: 'link';
        link: Link;
      };
    };
  };
  over: {
    id: string;
    data: {
      current?: {
        type: 'group' | 'link';
        group?: LinkGroup;
        link?: Link;
      };
    };
  } | null;
}
