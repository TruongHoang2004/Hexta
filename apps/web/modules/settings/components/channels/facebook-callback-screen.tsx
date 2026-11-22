"use client";

import Link from "next/link";
import { Notice, buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useFacebookOAuth } from "../../hooks/use-facebook-oauth";
import { OAuthPageOption } from "./oauth-page-option";

/** Meta OAuth landing page: pick which Pages to connect. Needs a Suspense boundary (search params). */
export function FacebookCallbackScreen() {
  const t = useT();
  const oauth = useFacebookOAuth();
  const { result, error, busy, selected } = oauth;
  if (!oauth.started) return <Notice error={t("oauth.notStarted")} />;
  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-semibold">{t("oauth.title")}</h1>
      <Notice error={error} />
      {!result && !error && <p role="status">{t("oauth.checking")}</p>}
      {error && (
        <Link href="/tenants" className="text-sm text-primary underline">
          {t("oauth.back")}
        </Link>
      )}
      {result && (
        <form
          onSubmit={(event) => {
            event.preventDefault();
            void oauth.connect();
          }}
          className="space-y-4 rounded-xl border border-border bg-surface p-5"
        >
          <p className="text-sm text-muted">
            {t("oauth.intro", { time: new Date(result.expires_at).toLocaleTimeString() })}
          </p>
          {!result.pages.length ? (
            <p className="text-sm">{t("oauth.noPages")}</p>
          ) : (
            <fieldset className="space-y-2">
              <legend className="sr-only">{t("oauth.pagesLegend")}</legend>
              {result.pages.map((page) => (
                <OAuthPageOption
                  key={page.page_id}
                  page={page}
                  checked={selected.includes(page.page_id)}
                  disabled={busy}
                  onChange={(checked) => oauth.toggle(page.page_id, checked)}
                />
              ))}
            </fieldset>
          )}
          <div className="flex flex-wrap gap-2">
            <button className={buttonClass} disabled={busy || !selected.length}>
              {busy ? t("oauth.connecting") : t("oauth.connect", { count: selected.length })}
            </button>
            <Link
              href={`/tenants/${result.tenant_id}/settings/channels`}
              className="rounded-lg border border-border px-3 py-2 text-sm hover:bg-surface"
            >
              {t("oauth.cancel")}
            </Link>
          </div>
        </form>
      )}
    </div>
  );
}
