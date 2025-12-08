export interface Operator {
  id: string;
  tenant_id: string;
  role: string;
  status: "AVAILABLE" | "OFFLINE";
  name?: string;
  created_at: string;
  updated_at: string;
}

export interface Inbox {
  id: string;
  tenant_id: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  id: string;
  tenant_id: string;
  inbox_id: string;
  external_id: string;
  customer_name?: string;
  customer_phone: string;
  state: "QUEUED" | "ALLOCATED" | "RESOLVED";
  priority: number;
  operator_id?: string;
  created_at: string;
  updated_at: string;
  allocated_at?: string;
  resolved_at?: string;
  last_message?: string;
}

export interface Label {
  id: string;
  tenant_id: string;
  inbox_id: string;
  name: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    total: number;
    per_page: number;
    cursor?: string;
    next_cursor?: string;
  };
}
