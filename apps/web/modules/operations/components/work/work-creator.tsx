"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { Field, buttonClass, inputClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import type { Order } from "@/modules/sales";
import { useRetryKey } from "../../hooks/use-retry-key";
import type { WorkflowDefinition, WorkInput } from "../../types";

interface Props {
  definitions: WorkflowDefinition[];
  orders: Order[];
  members: TenantMember[];
  busy: boolean;
  onCreate: (input: WorkInput, key: string) => Promise<unknown>;
}

/** Owner form: start a workflow on a confirmed order with a responsible employee. */
export function WorkCreator({ definitions, orders, members, busy, onCreate }: Props) {
  const t = useT();
  const person = useMemberName(members);
  const retryKey = useRetryKey();
  const [order, setOrder] = useState("");
  const [definition, setDefinition] = useState("");
  const [responsible, setResponsible] = useState("");
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (busy) return;
    const input = { order_id: order, workflow_version_id: definition, responsible_id: responsible };
    await onCreate(input, retryKey(input));
  }
  return (
    <details className="rounded-xl border border-border bg-surface p-4">
      <summary className="cursor-pointer font-semibold">{t("work.creator.title")}</summary>
      <p className="my-3 text-sm text-muted">{t("work.creator.body")}</p>
      <form onSubmit={(e) => void submit(e)}>
        <fieldset disabled={busy || !orders.length || !definitions.length} className="grid gap-3 md:grid-cols-3">
          <Field title={t("work.creator.order")}>
            <select className={inputClass} required value={order} onChange={(e) => setOrder(e.target.value)}>
              <option value="">{t("work.creator.selectOrder")}</option>
              {orders.map((row) => (
                <option key={row.id} value={row.id}>
                  {row.order_number}
                </option>
              ))}
            </select>
          </Field>
          <Field title={t("work.creator.workflow")}>
            <select className={inputClass} required value={definition} onChange={(e) => setDefinition(e.target.value)}>
              <option value="">{t("work.creator.selectWorkflow")}</option>
              {definitions.map((row) => (
                <option key={row.id} value={row.versions[0]?.id}>
                  {row.name} · v{row.version}
                </option>
              ))}
            </select>
          </Field>
          <Field title={t("work.creator.responsible")}>
            <select className={inputClass} required value={responsible} onChange={(e) => setResponsible(e.target.value)}>
              <option value="">{t("work.creator.selectEmployee")}</option>
              {members.map((row) => (
                <option key={row.id} value={row.user_id}>
                  {person(row.user_id)}
                </option>
              ))}
            </select>
          </Field>
          <div className="md:col-span-3">
            <button className={buttonClass} type="submit">
              {t("work.creator.submit")}
            </button>
          </div>
        </fieldset>
      </form>
      {!orders.length && <p className="mt-2 text-sm text-muted">{t("work.creator.noOrders")}</p>}
      {!definitions.length && <p className="mt-2 text-sm text-muted">{t("work.creator.noWorkflows")}</p>}
    </details>
  );
}
