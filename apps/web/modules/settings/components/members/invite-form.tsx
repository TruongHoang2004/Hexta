"use client";

import { useState } from "react";
import { Field, StatusChip, buttonClass, inputClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { InviteLink } from "./invite-link";

interface Props {
  busy: boolean;
  /** One-time link of the invitation just created, or "". */
  link: string;
  onInvite: (email: string) => Promise<boolean>;
}

/** Owner form: invite a member by email and show the one-time link. */
export function InviteForm({ busy, link, onInvite }: Props) {
  const t = useT();
  const [email, setEmail] = useState("");
  return (
    <form
      onSubmit={async (event) => {
        event.preventDefault();
        if (await onInvite(email)) setEmail("");
      }}
      className="space-y-3 rounded-xl border border-border bg-surface p-5"
    >
      <h2 className="font-semibold">{t("members.inviteTitle")}</h2>
      <div className="flex flex-wrap items-end gap-3">
        <div className="min-w-0 flex-[1_1_280px]">
          <Field title={t("members.email")}>
            <input
              className={inputClass}
              type="email"
              required
              maxLength={255}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={busy}
            />
          </Field>
        </div>
        <StatusChip>{t("members.roleNote")}</StatusChip>
        <button className={buttonClass} disabled={busy}>
          {busy ? t("members.creating") : t("members.create")}
        </button>
      </div>
      <p className="text-sm text-muted">{t("members.inviteHelp")}</p>
      {link && <InviteLink key={link} link={link} />}
    </form>
  );
}
