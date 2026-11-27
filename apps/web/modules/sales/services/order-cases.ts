import { sdk } from "@/lib/sdk";
import type { OrderCase, OrderCaseType, OrderCaseUpdateInput } from "../types";

const path = (tenant: string, order: string) => `/api/v1/tenants/${tenant}/orders/${order}/case`;

/** Cancellation/return case endpoints of one order. The API authorizes handler and owner actions. */
export const orderCases = {
  async get(tenant: string, order: string, signal?: AbortSignal): Promise<OrderCase | null> {
    const res = await sdk.client.get<{ data: { case: OrderCase | null } }>(path(tenant, order), { signal });
    return res.data.data.case;
  },
  async open(tenant: string, order: string, type: OrderCaseType): Promise<OrderCase> {
    const res = await sdk.client.post<{ data: OrderCase }>(path(tenant, order), { type }, {});
    return res.data.data;
  },
  async update(tenant: string, order: string, input: OrderCaseUpdateInput): Promise<OrderCase> {
    const res = await sdk.client.put<{ data: OrderCase }>(path(tenant, order), input, {});
    return res.data.data;
  },
  async sendBack(tenant: string, order: string, version: number, note: string): Promise<OrderCase> {
    const res = await sdk.client.post<{ data: OrderCase }>(`${path(tenant, order)}/send-back`, { version, note }, {});
    return res.data.data;
  },
  async finalize(tenant: string, order: string, version: number, note: string): Promise<OrderCase> {
    const res = await sdk.client.post<{ data: OrderCase }>(`${path(tenant, order)}/finalize`, { version, note }, {});
    return res.data.data;
  },
};
