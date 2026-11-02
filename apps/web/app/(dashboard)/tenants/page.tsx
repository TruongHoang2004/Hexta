"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/sales/workspace";
import { onboarding, type Application } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
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
  }, [revision]);

  async function refresh() {
    setRefreshing(true);
    try {
      await reload();
      setRevision((v) => v + 1);
      toast.success("Workspaces and permissions updated.");
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
        title="Organizations & Workspaces"
        description="Manage your active business workspaces or register a new brand organization."
        badge={
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
              <Sparkles className="w-3 h-3" /> Business Hub
            </span>
            {operator && (
              <Link
                href="/platform"
                className="inline-flex items-center gap-1 rounded-full bg-surface border border-border px-2.5 py-0.5 text-xs font-medium text-muted hover:text-foreground transition-colors"
              >
                <ShieldCheck className="w-3 h-3 text-primary" />
                Hexta Review Portal →
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
              <span>Refresh access</span>
            </Button>

            <Button size="sm" asChild className="h-9 gap-1.5 shadow-sm">
              <Link href="/tenants/register">
                <Plus className="w-4 h-4" />
                <span>Register Tenant</span>
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
              Your Workspaces
            </h2>
            <p className="text-xs text-muted">
              Select an active workspace to access orders, customers, and operations.
            </p>
          </div>
          <span className="text-xs font-medium text-muted bg-surface border border-border px-2.5 py-0.5 rounded-full">
            {tenants.length} {tenants.length === 1 ? "organization" : "organizations"}
          </span>
        </div>

        {tenants.length > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {tenants.map((row) =>
              row.status === "active" ? (
                <Link
                  key={row.id}
                  href={`/tenants/${row.id}/orders`}
                  className="group relative flex flex-col justify-between rounded-2xl border border-border bg-surface/70 p-5 shadow-xs transition-all duration-200 hover:border-primary/50 hover:bg-surface hover:shadow-md"
                >
                  <div>
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-white">
                        <Building2 className="w-5 h-5" />
                      </div>
                      <span className="inline-flex items-center gap-1 rounded-full bg-success/10 px-2 py-0.5 text-[11px] font-medium text-success">
                        <span className="h-1.5 w-1.5 rounded-full bg-success animate-pulse" />
                        Active
                      </span>
                    </div>

                    <h3 className="mt-4 text-sm font-semibold text-foreground group-hover:text-primary transition-colors truncate">
                      {row.name}
                    </h3>
                    <p className="mt-0.5 font-mono text-xs text-muted truncate">
                      {row.slug}
                    </p>
                  </div>

                  <div className="mt-5 flex items-center justify-between border-t border-border/50 pt-2.5 text-xs font-medium text-primary">
                    <span>Enter workspace</span>
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
                    <span className="rounded-full bg-muted/10 px-2 py-0.5 text-[11px] font-medium text-muted">
                      {row.status}
                    </span>
                  </div>
                  <div className="mt-4">
                    <h3 className="text-sm font-medium text-muted truncate">{row.name}</h3>
                    <p className="mt-0.5 text-xs text-muted truncate">Access unavailable</p>
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
                Register New Tenant
              </h3>
              <p className="mt-0.5 text-[11px] text-muted">
                Add another brand workspace
              </p>
            </Link>
          </div>
        ) : (
          <div className="relative overflow-hidden rounded-3xl border border-dashed border-border bg-surface/60 p-8 sm:p-10 text-center">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Store className="w-7 h-7" />
            </div>
            <h3 className="mt-4 text-base font-bold text-foreground">
              No active tenant workspace yet
            </h3>
            <p className="mx-auto mt-1.5 max-w-md text-xs text-muted leading-relaxed">
              Create an organization to start managing products, customer chats, and automated order fulfillment.
            </p>
            <div className="mt-5">
              <Button asChild className="h-9 px-5 gap-2 shadow-sm font-medium text-xs">
                <Link href="/tenants/register">
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>Register Your First Workspace</span>
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
                Registration Requests
              </h2>
              <p className="text-xs text-muted">
                Status of business organizations submitted for Hexta platform verification.
              </p>
            </div>
          </div>

          {loading ? (
            <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-surface/50 p-6 text-xs text-muted">
              <Loader2 className="w-4 h-4 animate-spin text-primary" />
              <span>Loading requests…</span>
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
                        <span className="inline-flex items-center gap-1 rounded-full bg-warning/15 px-2 py-0.5 text-[11px] font-semibold text-warning">
                          <Clock className="w-3 h-3" />
                          Pending Review
                        </span>
                      )}

                      {isApproved && (
                        <span className="inline-flex items-center gap-1 rounded-full bg-success/15 px-2 py-0.5 text-[11px] font-semibold text-success">
                          <CheckCircle2 className="w-3 h-3" />
                          Approved
                        </span>
                      )}

                      {isRejected && (
                        <span className="inline-flex items-center gap-1 rounded-full bg-error/15 px-2 py-0.5 text-[11px] font-semibold text-error">
                          <XCircle className="w-3 h-3" />
                          Declined
                        </span>
                      )}
                    </div>

                    {isPending && (
                      <p className="mt-2.5 text-xs text-muted leading-relaxed">
                        Your application is currently queued for review. Access will be automatically unlocked once approved.
                      </p>
                    )}

                    {isApproved && (
                      <div className="mt-2.5 flex items-center justify-between border-t border-success/20 pt-2">
                        <span className="text-xs text-success font-medium">Ready to access</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          asChild
                          className="h-7 text-xs text-success hover:bg-success/10 px-2"
                        >
                          <Link href={row.tenant_id ? `/tenants/${row.tenant_id}/orders` : "/tenants"}>
                            Enter Tenant →
                          </Link>
                        </Button>
                      </div>
                    )}

                    {isRejected && (
                      <div className="mt-2.5 space-y-2 border-t border-error/20 pt-2">
                        {row.review_reason && (
                          <p className="text-xs text-error font-medium">
                            Reason: {row.review_reason}
                          </p>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          asChild
                          className="h-7 text-xs text-error hover:bg-error/10 px-2 gap-1"
                        >
                          <Link href="/tenants/register">
                            <RotateCcw className="w-3 h-3" />
                            <span>Reapply with edits</span>
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
