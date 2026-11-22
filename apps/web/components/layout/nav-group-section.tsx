"use client";

import { ChevronDown } from "lucide-react";
import { useT } from "@/lib/i18n/provider";
import type { NavGroup } from "./nav-config";

/** A collapsible sidebar group. When collapsed over the current page, its title is highlighted. */
export function NavGroupSection({
  group,
  open,
  containsActive,
  onToggle,
  children,
}: {
  group: NavGroup;
  open: boolean;
  containsActive: boolean;
  onToggle: () => void;
  children: React.ReactNode;
}) {
  const t = useT();
  const Icon = group.icon;
  const panelID = `nav-group-${group.id}`;
  return (
    <div>
      <button
        type="button"
        aria-expanded={open}
        aria-controls={panelID}
        onClick={onToggle}
        className={`flex min-h-10 w-full items-center gap-2.5 rounded-lg px-3 text-sm font-medium hover:bg-background ${containsActive && !open ? "text-primary" : "text-foreground"}`}
      >
        <Icon aria-hidden className="h-4 w-4 shrink-0 text-muted" />
        <span className="flex-1 text-left">{t(group.label)}</span>
        <ChevronDown aria-hidden className={`h-4 w-4 shrink-0 text-muted transition-transform ${open ? "rotate-180" : ""}`} />
      </button>
      {open && (
        <div id={panelID} className="mb-1 ml-5 flex flex-col gap-0.5 border-l border-border pl-2">
          {children}
        </div>
      )}
    </div>
  );
}
