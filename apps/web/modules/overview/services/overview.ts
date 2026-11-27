import { sdk } from "@/lib/sdk";
import type { Overview } from "../types";

export interface OverviewQuery {
  /** RFC3339 instants; both or neither. */
  from?: string;
  to?: string;
  employee_id?: string;
}

export const overview = {
  /** Owner-only read; members receive 403. */
  async read(tenant: string, query: OverviewQuery, signal?: AbortSignal): Promise<Overview> {
    const res = await sdk.client.get<{ data: Overview }>(`/api/v1/tenants/${tenant}/overview`, {
      params: query,
      signal,
    });
    return res.data.data;
  },
};
