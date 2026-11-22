"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/layout/workspace-context";
import { onboarding, type Application } from "@/lib/services/onboarding";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import { StatusChip } from "@/components/primitives";
import { Button, toast } from "@hexta/ui";
import { PageHeader } from "@/components/page-header";
import {
  Building2,
  Sparkles,
  Clock,
  CheckCircle2,
  XCircle,
  ArrowRight,
  RefreshCw,
  ShieldCheck,
  Store,
  Loader2,
  RotateCcw,
  Plus,
} from "lucide-react";

export default function TenantsPage() {
  const { tenants, reload } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const userID = useAuthStore((state) => state.user?.id);
  const [requests, setRequests] = useState<Application[]>([]);
  const [operator, setOperator] = useState(false);

  // Status state
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    let ignore = false;
    async function load() {
      setLoading(true);
      try {
        const [r, o] = await Promise.all([
          onboarding.applications(),
          onboarding.operator(),
        ]);
        if (!ignore) {
          setRequests(r);
          setOperator(o);
        }
      } catch (err) {
        if (!ignore) {
          toast.error(errorText(err));
        }
      } finally {
        if (!ignore) setLoading(false);
      }
    }
    void load();
    return () => {
      ignore = true;
    };
  }, [revision, errorText]);

  async function refresh() {
    setRefreshing(true);
    try {
      await reload();
      setRevision((v) => v + 1);
      toast.success(t("workspaces.refreshed"));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      setRefreshing(false);
    }
  }

  return (
    <div className="space-y-8 pb-12">
      {/* Top Header via Unified PageHeader */}
      <PageHeader
        title={t("workspaces.title")}
        description={t("workspaces.description")}
        badge={
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
              <Sparkles className="w-3 h-3" /> {t("workspaces.hub")}
            </span>
            {operator && (
              <Link
                href="/platform"
                className="inline-flex items-center gap-1 rounded-full bg-surface border border-border px-2.5 py-0.5 text-xs font-medium text-muted hover:text-foreground transition-colors"
              >
                <ShieldCheck className="w-3 h-3 text-primary" />
                {t("workspaces.reviewPortal")}
              </Link>
            )}
          </div>
        }
        actions={
          <div className="flex items-center gap-2.5">
            <Button
              variant="outline"
              size="sm"
              disabled={refreshing || loading}
              onClick={() => void refresh()}
              className="h-9 gap-1.5"
            >
              <RefreshCw
                className={`w-3.5 h-3.5 ${refreshing ? "animate-spin" : ""}`}
              />
              <span>{t("workspaces.refresh")}</span>
            </Button>

            <Button size="sm" asChild className="h-9 gap-1.5 shadow-sm">
              <Link href="/tenants/register">
                <Plus className="w-4 h-4" />
                <span>{t("workspaces.register")}</span>
              </Link>
            </Button>
          </div>
        }
      />

      {/* Active Workspaces Section */}
      <section aria-labelledby="workspaces-heading" className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 id="workspaces-heading" className="text-base font-semibold text-foreground">
              {t("workspaces.yours")}
            </h2>
            <p className="text-xs text-muted">
              {t("workspaces.yoursHelp")}
            </p>
          </div>
          <span className="text-xs font-medium text-muted bg-surface border border-border px-2.5 py-0.5 rounded-full">
            {t("workspaces.count", { count: tenants.length })}
          </span>
        </div>

        {tenants.length > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {tenants.map((row) =>
              row.status === "active" ? (
                <Link
                  key={row.id}
                  href={`/tenants/${row.id}/inbox`}
                  className="group relative flex flex-col justify-between rounded-2xl border border-border bg-surface/70 p-5 shadow-xs transition-all duration-200 hover:border-primary/50 hover:bg-surface hover:shadow-md"
                >
                  <div>
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-white">
                        <Building2 className="w-5 h-5" />
                      </div>
                      <StatusChip tone="success">{t("workspaces.active")}</StatusChip>
                    </div>

                    <h3 className="mt-4 text-sm font-semibold text-foreground group-hover:text-primary transition-colors truncate">
                      {row.name}
                    </h3>
                    <p className="mt-0.5 font-mono text-xs text-muted truncate">
                      {row.slug}
                    </p>
                    <p className="mt-2">
                      <StatusChip tone={row.owner_id === userID ? "solid" : "neutral"}>
                        {row.owner_id === userID ? t("roles.owner") : t("roles.member")}
                      </StatusChip>
                    </p>
                  </div>

                  <div className="mt-5 flex items-center justify-between border-t border-border/50 pt-2.5 text-xs font-medium text-primary">
                    <span>{t("workspaces.enter")}</span>
                    <ArrowRight className="w-3.5 h-3.5 transition-transform duration-200 group-hover:translate-x-1" />
                  </div>
                </Link>
              ) : (
                <div
                  key={row.id}
                  className="flex flex-col justify-between rounded-2xl border border-dashed border-border bg-surface/40 p-5 opacity-70"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-muted/10 text-muted">
                      <Building2 className="w-5 h-5" />
                    </div>
                    <StatusChip>{row.status}</StatusChip>
                  </div>
                  <div className="mt-4">
                    <h3 className="text-sm font-medium text-muted truncate">{row.name}</h3>
                    <p className="mt-0.5 text-xs text-muted truncate">{t("workspaces.unavailable")}</p>
                  </div>
                </div>
              ),
            )}

            {/* Quick add workspace card */}
            <Link
              href="/tenants/register"
              className="flex min-h-[140px] flex-col items-center justify-center rounded-2xl border-2 border-dashed border-border/80 p-5 text-center transition-all duration-200 hover:border-primary/60 hover:bg-surface/60 group"
            >
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-muted/10 text-muted transition-colors group-hover:bg-primary/10 group-hover:text-primary">
                <Plus className="w-4.5 h-4.5" />
              </div>
              <h3 className="mt-2.5 text-xs font-semibold text-foreground group-hover:text-primary transition-colors">
                {t("workspaces.addTitle")}
              </h3>
              <p className="mt-0.5 text-[11px] text-muted">
                {t("workspaces.addBody")}
              </p>
            </Link>
          </div>
        ) : (
          <div className="relative overflow-hidden rounded-3xl border border-dashed border-border bg-surface/60 p-8 sm:p-10 text-center">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Store className="w-7 h-7" />
            </div>
            <h3 className="mt-4 text-base font-bold text-foreground">
              {t("workspaces.emptyTitle")}
            </h3>
            <p className="mx-auto mt-1.5 max-w-md text-xs text-muted leading-relaxed">
              {t("workspaces.emptyBody")}
            </p>
            <div className="mt-5">
              <Button asChild className="h-9 px-5 gap-2 shadow-sm font-medium text-xs">
                <Link href="/tenants/register">
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>{t("workspaces.emptyAction")}</span>
                  <ArrowRight className="w-3.5 h-3.5 ml-0.5" />
                </Link>
              </Button>
            </div>
          </div>
        )}
      </section>

      {/* Registration Requests Section */}
      {(loading || requests.length > 0) && (
        <section aria-labelledby="requests-heading" className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 id="requests-heading" className="text-base font-semibold text-foreground">
                {t("workspaces.requests")}
              </h2>
              <p className="text-xs text-muted">
                {t("workspaces.requestsHelp")}
              </p>
            </div>
          </div>

          {loading ? (
            <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-surface/50 p-6 text-xs text-muted">
              <Loader2 className="w-4 h-4 animate-spin text-primary" />
              <span>{t("workspaces.loadingRequests")}</span>
            </div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {requests.map((row) => {
                const isPending = row.status === "pending";
                const isApproved = row.status === "approved";
                const isRejected = row.status === "rejected";

                return (
                  <div
                    key={row.id}
                    className={`rounded-2xl border p-4 transition-colors ${
                      isPending
                        ? "border-warning/30 bg-warning/5"
                        : isApproved
                          ? "border-success/30 bg-success/5"
                          : "border-error/30 bg-error/5"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <h4 className="font-semibold text-foreground text-xs">{row.name}</h4>
                        <p className="font-mono text-[11px] text-muted mt-0.5">{row.slug}</p>
                      </div>

                      {isPending && (
                        <StatusChip tone="warning">
                          <Clock className="mr-1 h-3 w-3" aria-hidden="true" />
                          {t("workspaces.pending")}
                        </StatusChip>
                      )}

                      {isApproved && (
                        <StatusChip tone="success">
                          <CheckCircle2 className="mr-1 h-3 w-3" aria-hidden="true" />
                          {t("workspaces.approved")}
                        </StatusChip>
                      )}

                      {isRejected && (
                        <StatusChip tone="danger">
                          <XCircle className="mr-1 h-3 w-3" aria-hidden="true" />
                          {t("workspaces.rejected")}
                        </StatusChip>
                      )}
                    </div>

                    {isPending && (
                      <p className="mt-2.5 text-xs text-muted leading-relaxed">
                        {t("workspaces.pendingHelp")}
                      </p>
                    )}

                    {isApproved && (
                      <div className="mt-2.5 flex items-center justify-between border-t border-success/20 pt-2">
                        <span className="text-xs font-medium">{t("workspaces.ready")}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          asChild
                          className="h-7 px-2 text-xs"
                        >
                          <Link href={row.tenant_id ? `/tenants/${row.tenant_id}/inbox` : "/tenants"}>
                            {t("workspaces.enter")} →
                          </Link>
                        </Button>
                      </div>
                    )}

                    {isRejected && (
                      <div className="mt-2.5 space-y-2 border-t border-error/20 pt-2">
                        {row.review_reason && (
                          <p className="text-xs font-medium">
                            {t("workspaces.reason", { reason: row.review_reason })}
                          </p>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          asChild
                          className="h-7 gap-1 px-2 text-xs"
                        >
                          <Link href="/tenants/register">
                            <RotateCcw className="w-3 h-3" />
                            <span>{t("workspaces.reapply")}</span>
                          </Link>
                        </Button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </section>
      )}
    </div>
  );
}
