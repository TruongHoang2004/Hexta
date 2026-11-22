import type { ChipTone } from "@/components/primitives";
import type { Translate } from "@/lib/i18n/translate";
import type { CatalogRow, Customer, DraftInput, Order, OrderCheck, OrderLine, OrderListFilter, OrderListView, Product } from "./types";

const knownStatuses = ["draft", "confirmed", "processing", "fulfilled", "cancelled", "returned"] as const;
type KnownStatus = (typeof knownStatuses)[number];

export function statusChip(t: Translate, status: string): { label: string; tone: ChipTone } {
  const known = (knownStatuses as readonly string[]).includes(status) ? (status as KnownStatus) : null;
  return {
    label: known ? t(`orders.status.${known}`) : t("orders.status.other"),
    tone: status === "confirmed" ? "solid" : "neutral",
  };
}

export const orderListViews = ["all", "draft", "confirmed", "unassigned", "mine"] as const satisfies readonly OrderListView[];

export const viewFilters: Record<OrderListView, OrderListFilter> = {
  all: {},
  draft: { status: "draft" },
  confirmed: { status: "confirmed" },
  unassigned: { assignee: "unassigned" },
  mine: { assignee: "me" },
};

export const isCustomer = (row: CatalogRow): row is Customer => "full_name" in row;

export const blankCatalogForm = {
  full_name: "",
  phone: "",
  email: "",
  address: "",
  name: "",
  description: "",
  category: "",
  sku: "",
  variant_name: "",
  price: "",
};
export type CatalogForm = typeof blankCatalogForm;

export function catalogFormFrom(row: CatalogRow): CatalogForm {
  if (isCustomer(row))
    return {
      ...blankCatalogForm,
      full_name: row.full_name,
      phone: row.phone || "",
      email: row.email || "",
      address: row.address || "",
    };
  const variant = row.variants[0];
  return {
    ...blankCatalogForm,
    name: row.name,
    description: row.description || "",
    category: row.category || "",
    sku: variant?.sku || "",
    variant_name: variant?.variant_name || "",
    price: variant?.price_provided ? String(variant.price) : "",
  };
}

export function linesFrom(order: Order | null): OrderLine[] {
  return (
    order?.items.map((item) => ({
      key: item.id,
      variant_id: item.variant_id,
      quantity: String(item.quantity),
      unit_price: item.price_provided ? String(item.unit_price) : "",
    })) || []
  );
}

export function variantOptions(products: Product[]) {
  return products.flatMap((product) =>
    product.variants.map((variant) => ({
      ...variant,
      label: `${product.name} · ${variant.variant_name} (${variant.sku})`,
    })),
  );
}
export type VariantOption = ReturnType<typeof variantOptions>[number];

export interface DraftFields {
  customer: string;
  employee: string;
  delivery: boolean;
  address: string;
  notes: string;
  lines: OrderLine[];
}

export function draftInput(order: Order | null, fields: DraftFields, employeeID: string): DraftInput {
  return {
    version: order?.version || 0,
    customer_id: fields.customer || null,
    employee_id: employeeID || null,
    requires_delivery: fields.delivery,
    delivery_address: fields.address,
    notes: fields.notes,
    items: fields.lines.map((line) => ({
      variant_id: line.variant_id,
      quantity: Number(line.quantity),
      unit_price: line.unit_price.trim() === "" ? null : line.unit_price,
    })),
  };
}

/** Guidance mirrored from the server's confirmation gate; the server re-checks on confirm. */
export function orderChecks(
  t: Translate,
  { order, dirty, userID, customer, fields }: {
    order: Order | null;
    dirty: boolean;
    userID: string | undefined;
    customer: Customer | undefined;
    fields: DraftFields;
  },
): OrderCheck[] {
  const { lines, employee, delivery, address } = fields;
  const linePrices = lines.filter((line) => line.unit_price.trim() === "" || Number(line.unit_price) < 0);
  return [
    { key: "saved", passed: !!order && !dirty, label: t("orders.checklist.saved") },
    {
      key: "assignee",
      passed: !!userID && employee === userID,
      label: t("orders.checklist.assignee"),
      hint: employee && employee !== userID ? t("orders.checklist.assigneeOther") : undefined,
    },
    { key: "name", passed: !!(customer?.full_name ?? "").trim(), label: t("orders.checklist.customerName") },
    {
      key: "contact",
      passed: !!((customer?.phone ?? "").trim() || (customer?.email ?? "").trim()),
      label: t("orders.checklist.customerContact"),
    },
    {
      key: "items",
      passed:
        lines.length > 0 &&
        lines.every((line) => line.variant_id && Number.isInteger(Number(line.quantity)) && Number(line.quantity) > 0),
      label: t("orders.checklist.items"),
    },
    {
      key: "prices",
      passed: lines.length > 0 && linePrices.length === 0,
      label: t("orders.checklist.prices"),
      hint: linePrices.length ? t("orders.checklist.pricesMissing", { count: linePrices.length }) : undefined,
    },
    { key: "address", passed: !delivery || address.trim() !== "", label: t("orders.checklist.address") },
  ];
}
