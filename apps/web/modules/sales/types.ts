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
  /** Conversation session the draft was created from in Desk. */
  messaging_turn_id?: string;
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
  /** Conversation session a new draft is created from (Desk); ignored by the API on update. */
  messaging_turn_id?: string;
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

export type OrderCaseType = "cancellation" | "return";
export type OrderCaseStatus = "open" | "sent_back" | "finalized";
export type OrderCaseAction = "opened" | "updated" | "sent_back" | "finalized";

/** Returned quantity of one original order line; never above `original_quantity`. */
export interface OrderCaseLine {
  order_item_id: string;
  original_quantity: number;
  quantity: number;
}
export interface OrderCaseEvent {
  id: string;
  version: number;
  actor_id: string;
  action: OrderCaseAction;
  status: OrderCaseStatus;
  note: string;
  lines: OrderCaseLine[];
  created_at: string;
}
/** Cancellation/return case of a confirmed order; the original order is never changed by it. */
export interface OrderCase {
  id: string;
  order_id: string;
  type: OrderCaseType;
  status: OrderCaseStatus;
  handler_id: string;
  opened_by: string;
  progress_note: string;
  finalized_by: string | null;
  finalized_at: string | null;
  version: number;
  created_at: string;
  updated_at: string;
  lines: OrderCaseLine[];
  events: OrderCaseEvent[];
}
export interface OrderCaseUpdateInput {
  version: number;
  note: string;
  handler_id?: string;
  lines: { order_item_id: string; quantity: number }[];
}

/** Editable case fields; quantities are input strings keyed by order item ID. */
export interface OrderCaseForm {
  note: string;
  handler: string;
  quantities: Record<string, string>;
}
