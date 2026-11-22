import { sdk } from "@/lib/sdk";
import type { Customer, CustomerInput, DraftInput, Order, OrderListFilter, Page, Product, ProductInput } from "../types";

export const sales = {
  async orders(tenant: string, page: number, filter: OrderListFilter, signal?: AbortSignal): Promise<Page<Order>> {
    const res = await sdk.client.get<{ data: Page<Order> }>(`/api/v1/tenants/${tenant}/orders`, {
      params: { page, size: 20, ...filter },
      signal,
    });
    return res.data.data;
  },
  async list<T>(
    resource: "customers" | "products" | "orders",
    tenant: string,
    page = 1,
    signal?: AbortSignal,
  ): Promise<Page<T>> {
    const res = await sdk.client.get<{ data: Page<T> }>(`/api/v1/tenants/${tenant}/${resource}`, {
      params: { page, size: 20 },
      signal,
    });
    return res.data.data;
  },
  async all<T>(resource: "customers" | "products" | "orders", tenant: string, signal?: AbortSignal): Promise<T[]> {
    const rows: T[] = [];
    for (let page = 1; ; page++) {
      const res = await sdk.client.get<{ data: Page<T> }>(`/api/v1/tenants/${tenant}/${resource}`, {
        params: { page, size: 100 },
        signal,
      });
      rows.push(...res.data.data.items);
      if (rows.length >= res.data.data.total || !res.data.data.items.length) return rows;
    }
  },
  async saveCustomer(tenant: string, id: string | null, input: CustomerInput): Promise<Customer> {
    const res = await sdk.client.request<{ data: Customer }>({
      method: id ? "PUT" : "POST",
      url: `/api/v1/tenants/${tenant}/customers${id ? `/${id}` : ""}`,
      data: input,
    });
    return res.data.data;
  },
  async saveProduct(tenant: string, id: string | null, input: ProductInput): Promise<Product> {
    const res = await sdk.client.request<{ data: Product }>({
      method: id ? "PUT" : "POST",
      url: `/api/v1/tenants/${tenant}/products${id ? `/${id}` : ""}`,
      data: input,
    });
    return res.data.data;
  },
  async archive(resource: "customers" | "products", tenant: string, id: string): Promise<void> {
    await sdk.client.delete(`/api/v1/tenants/${tenant}/${resource}/${id}`, {});
  },
  async order(tenant: string, id: string, signal?: AbortSignal): Promise<Order> {
    const res = await sdk.client.get<{ data: Order }>(`/api/v1/tenants/${tenant}/orders/${id}`, { signal });
    return res.data.data;
  },
  async saveDraft(tenant: string, id: string | null, input: DraftInput): Promise<Order> {
    const res = await sdk.client.request<{ data: Order }>({
      method: id ? "PUT" : "POST",
      url: `/api/v1/tenants/${tenant}/orders${id ? `/${id}` : ""}`,
      data: input,
    });
    return res.data.data;
  },
  async confirm(tenant: string, id: string, version: number): Promise<Order> {
    const res = await sdk.client.post<{ data: Order }>(
      `/api/v1/tenants/${tenant}/orders/${id}/confirm`,
      { version },
      {},
    );
    return res.data.data;
  },
};
