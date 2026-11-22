"use client";

import { useT } from "@/lib/i18n/provider";
import type { SelectablePage } from "../../types";

interface Props {
  page: SelectablePage;
  checked: boolean;
  disabled: boolean;
  onChange: (checked: boolean) => void;
}

/** One Page from the OAuth result; Pages owned by another workspace cannot be picked. */
export function OAuthPageOption({ page, checked, disabled, onChange }: Props) {
  const t = useT();
  const taken = page.availability === "owned_by_other_tenant";
  return (
    <label className={`flex items-start gap-3 rounded-lg border border-border p-3 ${taken ? "opacity-60" : ""}`}>
      <input
        type="checkbox"
        className="mt-1"
        disabled={taken || disabled}
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span>
        <span className="block font-medium">{page.page_name}</span>
        <span className="block text-xs text-muted">{t(`oauth.availability.${page.availability}`)}</span>
      </span>
    </label>
  );
}
