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
export interface OrderListFilter {
  status?: "draft" | "confirmed";
  assignee?: "me" | "unassigned";
}

export type CatalogKind = "customers" | "products";
export type CatalogRow = Customer | Product;

/** Order list views offered in the UI; each maps to a server-side filter. */
export type OrderListView = "all" | "draft" | "confirmed" | "unassigned" | "mine";

/** One editable order line; strings mirror the form inputs. */
export interface OrderLine {
  key: string;
  variant_id: string;
  quantity: string;
  unit_price: string;
}

export interface OrderCheck {
  key: string;
  passed: boolean;
  label: string;
  hint?: string;
}
