"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import { useT } from "@/lib/i18n/provider";
import { navGroups } from "./nav-config";
import { NavGroupSection } from "./nav-group-section";
import { NavLink } from "./nav-link";

/** Grouped workspace navigation. The group holding the current page opens whenever navigation enters it. */
export function SidebarNav({ tenantID }: { tenantID: string }) {
  const t = useT();
  const pathname = usePathname();
  const base = `/tenants/${tenantID}`;
  const isActive = (path: string) => pathname === `${base}/${path}` || pathname.startsWith(`${base}/${path}/`);
  const activeGroup = navGroups.find((group) => group.items.some((item) => isActive(item.path)))?.id ?? null;
  const [openGroups, setOpenGroups] = useState<string[]>([]);
  const [seenActiveGroup, setSeenActiveGroup] = useState<string | null>(null);
  if (activeGroup !== seenActiveGroup) {
    setSeenActiveGroup(activeGroup);
    if (activeGroup && !openGroups.includes(activeGroup)) setOpenGroups([...openGroups, activeGroup]);
  }
  const toggle = (id: string) =>
    setOpenGroups((current) => (current.includes(id) ? current.filter((g) => g !== id) : [...current, id]));
  return (
    <nav aria-label={t("shell.navLabel")} className="flex flex-1 flex-col gap-1 overflow-y-auto p-3">
      {navGroups.map((group) => (
        <NavGroupSection
          key={group.id}
          group={group}
          open={openGroups.includes(group.id)}
          containsActive={group.id === activeGroup}
          onToggle={() => toggle(group.id)}
        >
          {group.items.map((item) => (
            <NavLink key={item.path} href={`${base}/${item.path}`} active={isActive(item.path)}>
              {t(item.label)}
            </NavLink>
          ))}
        </NavGroupSection>
      ))}
    </nav>
  );
}
