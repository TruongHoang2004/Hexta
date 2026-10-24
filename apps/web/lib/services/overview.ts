import { sdk } from "@/lib/sdk";
import type { Page } from "./sales";

export interface OverviewWorkRow {
  id: string;
  order_id: string;
  order_number: string;
  workflow_version_id: string;
  workflow_name: string;
  workflow_version_number: number;
  current_step: number;
  current_step_name: string;
  responsible_id: string | null;
  recipient_id: string | null;
  state: "active" | "waiting" | "completed";
  version: number;
  created_at: string;
  updated_at: string;
}
export interface OverviewResponse {
  filters: {
    from: string | null;
    to: string | null;
    employee_id: string | null;
    assignment: "all" | "assigned" | "unassigned";
  };
  orders: {
    total: number;
    draft: number;
    confirmed: number;
    processing: number;
    fulfilled: number;
    cancelled: number;
    returned: number;
    other: number;
  };
  work: {
    total: number;
    active: number;
    waiting: number;
    completed: number;
    unassigned: number;
  };
  work_items: Page<OverviewWorkRow>;
}
export async function getOverview(
  tenant: string,
  query: string,
  signal?: AbortSignal,
): Promise<OverviewResponse> {
  const response = await sdk.client.get<{ data: OverviewResponse }>(
    `/api/v1/tenants/${tenant}/overview${query ? `?${query}` : ""}`,
    { signal },
  );
  return response.data.data;
}
