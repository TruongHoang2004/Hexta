"use client";

import { createContext, useCallback, useContext, useMemo } from "react";
import { useRouter } from "next/navigation";
import { localeCookie, type Locale } from "./config";
import { translate, type MessageKey, type MessageParams, type Translate } from "./translate";

interface I18nState {
  locale: Locale;
  t: Translate;
}

const I18nContext = createContext<I18nState | null>(null);

export function I18nProvider({ locale, children }: { locale: Locale; children: React.ReactNode }) {
  const value = useMemo<I18nState>(
    () => ({
      locale,
      t: (key: MessageKey, params?: MessageParams) => translate(locale, key, params),
    }),
    [locale],
  );
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

function useI18n(): I18nState {
  const state = useContext(I18nContext);
  if (!state) throw new Error("I18nProvider is required");
  return state;
}

export function useT(): Translate {
  return useI18n().t;
}

export function useLocale(): Locale {
  return useI18n().locale;
}

/** Persists the locale in a cookie and re-renders server components with it. */
export function useSetLocale(): (locale: Locale) => void {
  const router = useRouter();
  return useCallback(
    (locale: Locale) => {
      document.cookie = `${localeCookie}=${locale}; path=/; max-age=31536000; samesite=lax`;
      router.refresh();
    },
    [router],
  );
}
