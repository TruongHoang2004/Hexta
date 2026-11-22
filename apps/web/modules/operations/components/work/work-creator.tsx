"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { Field, Picker, buttonClass } from "@/components/primitives";
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
            <Picker
              required
              value={order}
              onChange={setOrder}
              placeholder={t("work.creator.selectOrder")}
              options={orders.map((row) => ({ value: row.id, label: row.order_number }))}
            />
          </Field>
          <Field title={t("work.creator.workflow")}>
            <Picker
              required
              value={definition}
              onChange={setDefinition}
              placeholder={t("work.creator.selectWorkflow")}
              options={definitions
                .filter((row) => row.versions[0])
                .map((row) => ({ value: row.versions[0].id, label: `${row.name} · v${row.version}` }))}
            />
          </Field>
          <Field title={t("work.creator.responsible")}>
            <Picker
              required
              value={responsible}
              onChange={setResponsible}
              placeholder={t("work.creator.selectEmployee")}
              options={members.map((row) => ({ value: row.user_id, label: person(row.user_id) }))}
            />
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
