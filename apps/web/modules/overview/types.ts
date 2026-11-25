/** Period presets; ranges are computed in the browser's local time zone. */
export type OverviewPeriod = "today" | "7d" | "30d" | "all";

export interface OverviewFilter {
  period: OverviewPeriod;
  /** "" means every employee. */
  employeeID: string;
}

export interface OverviewOrderBucket {
  /** Stored order status, or "other" for statuses the API does not report separately. */
  status: string;
  count: number;
  order_value: string;
}

export interface OverviewStep {
  workflow_version_id: string;
  workflow_name: string;
  workflow_version_number: number;
  /** Zero-based position in the pinned workflow version. */
  step_ordinal: number;
  step_name: string;
  active: number;
  waiting: number;
  completed: number;
  /** Subset of active. */
  unassigned: number;
}

export interface OverviewStuckItem {
  work_item_id: string;
  order_id: string;
  order_number: string;
  workflow_name: string;
  step_ordinal: number;
  step_name: string;
  state: "active" | "waiting";
  responsible_id: string | null;
  recipient_id: string | null;
  reason: "unassigned" | "waiting_too_long";
  waiting_since: string;
}

export interface Overview {
  filters: { from: string | null; to: string | null; employee_id: string | null };
  /** False only when the workspace has no orders and no work at all. */
  has_activity: boolean;
  orders: { total: number; order_value: string; by_status: OverviewOrderBucket[] };
  work: {
    total: number;
    active: number;
    waiting: number;
    completed: number;
    unassigned: number;
    by_steps: OverviewStep[];
  };
  stuck: { threshold_hours: number; total: number; items: OverviewStuckItem[] };
}
