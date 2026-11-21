"use client";
import { useWorkspace } from "@/components/sales/workspace";
import { useT } from "@/lib/i18n/provider";
import { StatusChip } from "@/components/sales/primitives";

export default function SubscriptionPage() {
  const { tenant } = useWorkspace();
  const t = useT();
  return (
    <section className="max-w-xl space-y-4 rounded-xl border border-border bg-surface p-6">
      <h2 className="text-lg font-semibold">{t("settingsPages.planTitle")}</h2>
      <p className="text-muted">{t("settingsPages.planIntro", { name: tenant?.name ?? "" })}</p>
      <dl className="space-y-3">
        <div className="flex justify-between">
          <dt>{t("settingsPages.plan")}</dt>
          <dd className="font-semibold uppercase">{tenant?.plan}</dd>
        </div>
        <div className="flex justify-between">
          <dt>{t("settingsPages.access")}</dt>
          <dd>
            <StatusChip tone={tenant?.status === "active" ? "success" : "neutral"}>{tenant?.status}</StatusChip>
          </dd>
        </div>
      </dl>
      <p className="text-sm text-muted">{t("settingsPages.planNote")}</p>
    </section>
  );
}
