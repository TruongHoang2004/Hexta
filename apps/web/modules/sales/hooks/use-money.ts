"use client";

import { useLocale } from "@/lib/i18n/provider";

/** Formats an amount in VND for the current locale; non-numeric values pass through. */
export function useMoney() {
  const locale = useLocale();
  return (value: string | number) => {
    const amount = Number(value);
    return Number.isFinite(amount)
      ? `${new Intl.NumberFormat(locale === "vi" ? "vi-VN" : "en-US", { maximumFractionDigits: 2 }).format(amount)} đ`
      : String(value);
  };
}
