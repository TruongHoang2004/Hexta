import { ClipboardList, LayoutDashboard, MessagesSquare, ScrollText, Settings, ShoppingBag, type LucideIcon } from "lucide-react";
import type { MessageKey } from "@/lib/i18n/translate";

export interface NavItem {
  /** Path below `/tenants/:tenantId/`. */
  path: string;
  label: MessageKey;
}

/** A sidebar group; one per frontend domain module. */
export interface NavGroup {
  id: string;
  label: MessageKey;
  icon: LucideIcon;
  items: NavItem[];
  /** Hidden for members. Navigation only; the API still authorizes every request. */
  ownerOnly?: boolean;
}

export const settingsNav: NavItem[] = [
  { path: "settings/members", label: "nav.members" },
  { path: "settings/channels", label: "nav.channels" },
  { path: "settings/reply-mode", label: "nav.replyMode" },
  { path: "settings/knowledge", label: "nav.knowledge" },
  { path: "settings/script-permissions", label: "nav.scriptPermissions" },
  { path: "settings/subscription", label: "nav.plan" },
];

export const navGroups: NavGroup[] = [
  {
    id: "overview",
    label: "nav.groups.overview",
    icon: LayoutDashboard,
    ownerOnly: true,
    items: [{ path: "overview", label: "nav.overview" }],
  },
  {
    id: "conversations",
    label: "nav.groups.conversations",
    icon: MessagesSquare,
    items: [
      { path: "inbox", label: "nav.inbox" },
      { path: "storefront", label: "nav.consultant" },
    ],
  },
  {
    id: "sales",
    label: "nav.groups.sales",
    icon: ShoppingBag,
    items: [
      { path: "orders", label: "nav.orders" },
      { path: "customers", label: "nav.customers" },
      { path: "products", label: "nav.products" },
    ],
  },
  {
    id: "operations",
    label: "nav.groups.operations",
    icon: ClipboardList,
    items: [
      { path: "work", label: "nav.work" },
      { path: "workflows", label: "nav.workflows" },
    ],
  },
  {
    id: "scripts",
    label: "nav.groups.scripts",
    icon: ScrollText,
    items: [{ path: "scripts", label: "nav.scripts" }],
  },
  { id: "settings", label: "shell.settings", icon: Settings, items: settingsNav },
];
