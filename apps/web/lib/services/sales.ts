import { sdk } from "@/lib/sdk";

export interface Customer {
  id: string;
  full_name: string;
  phone: string;
  email: string;
  address: string;
}
export interface Variant {
  id: string;
  sku: string;
  variant_name: string;
  price: string | number;
  price_provided: boolean;
}
export interface Product {
  id: string;
  name: string;
  description: string;
  category: string;
  variants: Variant[];
}
export interface OrderItem {
  id: string;
  variant_id: string;
  quantity: number;
  unit_price: string | number;
  price_provided: boolean;
}
export interface Order {
  id: string;
  order_number: string;
  customer_id?: string;
  employee_id?: string;
  status: string;
  version: number;
  requires_delivery: boolean;
  delivery_address: string;
  notes: string;
  items: OrderItem[];
  subtotal: string | number;
  discount: string | number;
  total_amount: string | number;
}
export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}
export interface ProductInput {
  name: string;
  description: string;
  category: string;
  sku: string;
  variant_name: string;
  price: string | null;
}
export interface DraftInput {
  version: number;
  customer_id: string | null;
  employee_id: string | null;
  requires_delivery: boolean;
  delivery_address: string;
  notes: string;
  items: { variant_id: string; quantity: number; unit_price: string | null }[];
}
export type CustomerInput = Omit<Customer, "id">;
export const sales = {
  async list<T>(
    resource: "customers" | "products" | "orders",
    tenant: string,
    page = 1,
    signal?: AbortSignal,
  ): Promise<Page<T>> {
    const res = await sdk.client.get<{ data: Page<T> }>(
      `/api/v1/tenants/${tenant}/${resource}`,
      {
        params: { page, size: 20 },
        signal,
      },
    );
    return res.data.data;
  },
  async all<T>(
    resource: "customers" | "products",
    tenant: string,
    signal?: AbortSignal,
  ): Promise<T[]> {
    const rows: T[] = [];
    for (let page = 1; ; page++) {
      const res = await sdk.client.get<{ data: Page<T> }>(
        `/api/v1/tenants/${tenant}/${resource}`,
        { params: { page, size: 100 }, signal },
      );
      rows.push(...res.data.data.items);
      if (rows.length >= res.data.data.total || !res.data.data.items.length)
        return rows;
    }
  },
  async saveCustomer(
    tenant: string,
    id: string | null,
    input: CustomerInput,
  ): Promise<Customer> {
    const res = await sdk.client.request<{ data: Customer }>({
      method: id ? "PUT" : "POST",
      url: `/api/v1/tenants/${tenant}/customers${id ? `/${id}` : ""}`,
      data: input,
    });
    return res.data.data;
  },
  async saveProduct(
    tenant: string,
    id: string | null,
    input: ProductInput,
  ): Promise<Product> {
    const res = await sdk.client.request<{ data: Product }>({
      method: id ? "PUT" : "POST",
      url: `/api/v1/tenants/${tenant}/products${id ? `/${id}` : ""}`,
      data: input,
    });
    return res.data.data;
  },
  async archive(
    resource: "customers" | "products",
    tenant: string,
    id: string,
  ): Promise<void> {
    await sdk.client.delete(`/api/v1/tenants/${tenant}/${resource}/${id}`, {});
  },
  async order(
    tenant: string,
    id: string,
    signal?: AbortSignal,
  ): Promise<Order> {
    const res = await sdk.client.get<{ data: Order }>(
      `/api/v1/tenants/${tenant}/orders/${id}`,
      {
        signal,
      },
    );
    return res.data.data;
  },
  async saveDraft(
    tenant: string,
    id: string | null,
    input: DraftInput,
  ): Promise<Order> {
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
export const errorText = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "Unable to complete this action. Please try again.";
