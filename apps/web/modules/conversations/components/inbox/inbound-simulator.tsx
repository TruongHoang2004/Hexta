"use client";

import { Textarea } from "@hexta/ui";
import { inputClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useInboundSimulator } from "../../hooks/use-inbound-simulator";

/** Collapsible form that adds a simulated customer message to the desk. */
export function InboundSimulator({
  tenantID,
  onSaved,
}: {
  tenantID: string;
  onSaved: (threadID: string) => void;
}) {
  const t = useT();
  const form = useInboundSimulator(tenantID, onSaved);
  return (
    <details className="rounded-xl border border-border bg-surface p-4">
      <summary className="cursor-pointer font-medium">{t("desk.simulator.title")}</summary>
      <p className="my-3 text-sm text-muted">{t("desk.simulator.help")}</p>
      <form
        onSubmit={(event) => {
          event.preventDefault();
          void form.submit();
        }}
        className="grid gap-3 md:grid-cols-2"
      >
        <label className="space-y-1.5 text-sm">
          {t("desk.simulator.reference")}
          <input
            required
            maxLength={120}
            disabled={form.busy}
            value={form.external}
            onChange={(e) => form.setExternal(e.target.value)}
            className={inputClass}
          />
        </label>
        <label className="space-y-1.5 text-sm">
          {t("desk.simulator.name")}
          <input
            required
            maxLength={120}
            disabled={form.busy}
            value={form.name}
            onChange={(e) => form.setName(e.target.value)}
            className={inputClass}
          />
        </label>
        <label className="space-y-1.5 text-sm md:col-span-2">
          {t("desk.simulator.message")}
          <Textarea
            required
            maxLength={4000}
            disabled={form.busy}
            value={form.body}
            onChange={(e) => form.setBody(e.target.value)}
          />
        </label>
        {form.error && (
          <p role="alert" className="text-sm text-error md:col-span-2">
            {form.error}
          </p>
        )}
        <div className="md:col-span-2">
          <button className={secondaryClass} disabled={form.busy || !form.ready}>
            {form.busy ? t("desk.simulator.saving") : t("desk.simulator.submit")}
          </button>
        </div>
      </form>
    </details>
  );
}
