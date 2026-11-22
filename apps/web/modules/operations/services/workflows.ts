import { sdk } from "@/lib/sdk";
import type { Page } from "@/modules/sales";
import type {
  WorkCommand,
  WorkCommandFields,
  WorkflowDefinition,
  WorkflowInput,
  WorkInput,
  WorkItem,
  WorkView,
} from "../types";

const base = (tenant: string) => `/api/v1/tenants/${tenant}`;

export const workflows = {
  async definitions(tenant: string, page = 1, signal?: AbortSignal): Promise<Page<WorkflowDefinition>> {
    const res = await sdk.client.get<{ data: Page<WorkflowDefinition> }>(`${base(tenant)}/workflow-definitions`, {
      params: { page, size: 20 },
      signal,
    });
    return res.data.data;
  },
  async allDefinitions(tenant: string, signal?: AbortSignal): Promise<WorkflowDefinition[]> {
    const rows: WorkflowDefinition[] = [];
    for (let page = 1; ; page++) {
      const result = await this.definitions(tenant, page, signal);
      rows.push(...result.items);
      if (!result.items.length || rows.length >= result.total) return rows;
    }
  },
  async publish(tenant: string, id: string | null, input: WorkflowInput, key: string): Promise<WorkflowDefinition> {
    const res = await sdk.client.request<{ data: WorkflowDefinition }>({
      method: "POST",
      url: `${base(tenant)}/workflow-definitions${id ? `/${id}/versions` : ""}`,
      data: input,
      headers: { "Idempotency-Key": key },
    });
    return res.data.data;
  },
  async items(tenant: string, page = 1, signal?: AbortSignal, view: WorkView = "all"): Promise<Page<WorkItem>> {
    const res = await sdk.client.get<{ data: Page<WorkItem> }>(`${base(tenant)}/work-items`, {
      params: { page, size: 20, view },
      signal,
    });
    return res.data.data;
  },
  async item(tenant: string, id: string): Promise<WorkItem> {
    const res = await sdk.client.get<{ data: WorkItem }>(`${base(tenant)}/work-items/${id}`, {});
    return res.data.data;
  },
  async create(tenant: string, input: WorkInput, key: string): Promise<WorkItem> {
    const res = await sdk.client.request<{ data: WorkItem }>({
      method: "POST",
      url: `${base(tenant)}/work-items`,
      data: input,
      headers: { "Idempotency-Key": key },
    });
    return res.data.data;
  },
  async command(
    tenant: string,
    item: WorkItem,
    action: WorkCommand,
    fields: WorkCommandFields = {},
  ): Promise<WorkItem> {
    const res = await sdk.client.post<{ data: WorkItem }>(
      `${base(tenant)}/work-items/${item.id}/${action}`,
      { expected_version: item.version, ...fields },
      {},
    );
    return res.data.data;
  },
};
