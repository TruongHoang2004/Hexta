"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Check, ChevronsUpDown } from "lucide-react";
import type { Tenant } from "@hexta/sdk";
import { useT } from "@/lib/i18n/provider";

/** Current workspace name and role; opens a list of the user's workspaces to switch to. */
export function WorkspaceSwitcher({
  current,
  tenants,
  roleLabel,
}: {
  current: Tenant;
  tenants: Tenant[];
  roleLabel: string;
}) {
  const t = useT();
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (!open) return;
    function onPointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    }
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
        buttonRef.current?.focus();
      }
    }
    document.addEventListener("pointerdown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);
  return (
    <div ref={rootRef} className="relative border-b border-border p-3">
      <button
        ref={buttonRef}
        type="button"
        aria-expanded={open}
        aria-controls="workspace-switcher"
        aria-label={`${t("shell.changeWorkspace")}: ${current.name}`}
        onClick={() => setOpen((value) => !value)}
        className="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-background"
      >
        <span className="min-w-0 flex-1">
          <span className="block truncate font-semibold">{current.name}</span>
          <span className="block text-xs text-muted">{roleLabel}</span>
        </span>
        <ChevronsUpDown aria-hidden className="h-4 w-4 shrink-0 text-muted" />
      </button>
      {open && (
        <ul
          id="workspace-switcher"
          className="absolute inset-x-3 top-full z-20 mt-1 max-h-72 overflow-y-auto rounded-lg border border-border bg-surface p-1 shadow-lg"
        >
          {tenants.map((row) => {
            const selected = row.id === current.id;
            return (
              <li key={row.id}>
                <Link
                  href={`/tenants/${row.id}/inbox`}
                  aria-current={selected ? "true" : undefined}
                  onClick={() => setOpen(false)}
                  className={`flex min-h-9 items-center gap-2 rounded-md px-2 text-sm hover:bg-background ${selected ? "font-semibold" : ""}`}
                >
                  <span className="min-w-0 flex-1 truncate">{row.name}</span>
                  {selected && <Check aria-hidden className="h-4 w-4 shrink-0 text-primary" />}
                </Link>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
