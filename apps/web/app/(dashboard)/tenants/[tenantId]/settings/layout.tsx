"use client";

import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import { settingsNav } from "@/components/sales/workspace";
import { useT } from "@/lib/i18n/provider";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const tenantID = typeof params.tenantId === "string" ? params.tenantId : "";
  const pathname = usePathname();
  const t = useT();
  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-semibold">{t("shell.settings")}</h1>
      <nav aria-label={t("shell.settingsTabs")} className="flex flex-wrap gap-1 border-b border-border">
        {settingsNav.map((item) => {
          const href = `/tenants/${tenantID}/${item.path}`;
          const active = pathname === href;
          return (
            <Link
              key={item.path}
              href={href}
              aria-current={active ? "page" : undefined}
              className={`-mb-px inline-flex min-h-11 items-center border-b-2 px-3 text-sm ${active ? "border-foreground font-semibold" : "border-transparent text-muted hover:text-foreground"}`}
            >
              {t(item.label)}
            </Link>
          );
        })}
      </nav>
      {children}
    </div>
  );
}
