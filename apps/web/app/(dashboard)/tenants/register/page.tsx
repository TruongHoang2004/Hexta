"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { onboarding } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
import { Button, Input, Label, toast } from "@hexta/ui";
import { PageHeader } from "@/components/page-header";
import {
  Store,
  Sparkles,
  ArrowRight,
  Globe,
  ShieldCheck,
  Loader2,
  RotateCcw,
  Bot,
  MessageSquare,
  Workflow,
  CheckCircle2,
} from "lucide-react";

function toSlug(value: string): string {
  return value
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "") // remove diacritics
    .replace(/[đĐ]/g, "d")
    .replace(/[^a-z0-9\s-]/g, "") // remove special chars
    .trim()
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .slice(0, 100);
}

export default function RegisterTenantPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugCustomized, setSlugCustomized] = useState(false);
  const [busy, setBusy] = useState(false);

  const handleNameChange = (val: string) => {
    setName(val);
    if (!slugCustomized) {
      setSlug(toSlug(val));
    }
  };

  const handleSlugChange = (val: string) => {
    const sanitized = val.toLowerCase().replace(/[^a-z0-9-]/g, "");
    setSlug(sanitized);
    setSlugCustomized(true);
  };

  const handleResetSlug = () => {
    setSlug(toSlug(name));
    setSlugCustomized(false);
  };

  async function apply(event: React.FormEvent) {
    event.preventDefault();
    const trimmedName = name.trim();
    const trimmedSlug = slug.trim();

    if (trimmedName.length < 2) {
      toast.error("Business name must be at least 2 characters.");
      return;
    }

    if (!trimmedSlug || !/^[a-z0-9]+(-[a-z0-9]+)*$/.test(trimmedSlug)) {
      toast.error(
        "Slug must be lowercase alphanumeric with optional hyphens (e.g. 'my-shop').",
      );
      return;
    }

    setBusy(true);
    try {
      await onboarding.apply(trimmedName, trimmedSlug);
      toast.success(
        "Application submitted! Hexta operators will review and approve access soon.",
      );
      router.push("/tenants");
    } catch (err) {
      toast.error(errorText(err));
      setBusy(false);
    }
  }

  const isValidSlug = slug.length >= 2 && /^[a-z0-9]+(-[a-z0-9]+)*$/.test(slug);

  return (
    <div className="space-y-6 pb-12">
      {/* Unified Compact PageHeader with integrated Breadcrumb/Back link */}
      <PageHeader
        backHref="/tenants"
        backLabel="Workspaces"
        title="Register a Business Tenant"
        description="Create an isolated organization workspace to manage products, orders, customers, and AI copilot agents."
        badge={
          <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
            <Sparkles className="w-3 h-3" /> Setup
          </span>
        }
      />

      <div className="grid gap-8 lg:grid-cols-12 items-start pt-1">
        {/* Left Side: Value Proposition & Onboarding Guide */}
        <div className="lg:col-span-5 space-y-5">
          <div>
            <h2 className="text-sm font-semibold text-foreground">
              What&apos;s included in your workspace
            </h2>
            <p className="mt-0.5 text-xs text-muted leading-relaxed">
              Every Hexta tenant comes preconfigured with enterprise-grade capabilities.
            </p>
          </div>

          <div className="space-y-3">
            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <MessageSquare className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  Omni-channel Inbox
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  Consolidate conversations from social media, chat widgets, and direct messages in one live inbox.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Bot className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  AI Sales Copilot
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  Deliver instant product recommendations, order inquiries, and autonomous conversational checkout.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Workflow className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  Configurable Workflows
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  Define automated fulfillment pipelines, human escalation rules, and inventory alerts.
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-start gap-3 rounded-2xl border border-border/60 bg-muted/5 p-3.5 text-xs text-muted">
            <ShieldCheck className="w-4 h-4 text-primary shrink-0 mt-0.5" />
            <div className="leading-relaxed text-[11px]">
              <span className="font-semibold text-foreground">Transparent Review: </span>
              Registration requests are reviewed by Hexta platform operators to maintain workspace isolation and reliability. Your user account automatically receives Owner privileges upon approval.
            </div>
          </div>
        </div>

        {/* Right Side: The Registration Form */}
        <div className="lg:col-span-7">
          <div className="relative overflow-hidden rounded-3xl border border-border bg-surface/90 p-5 sm:p-7 shadow-xs backdrop-blur">
            {/* Ambient lighting effect */}
            <div className="absolute top-0 right-0 -z-10 h-64 w-64 rounded-full bg-primary/10 blur-3xl pointer-events-none" />

            <div className="flex items-center gap-3 border-b border-border/60 pb-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Store className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-base font-bold text-foreground">
                  Organization Information
                </h2>
                <p className="text-xs text-muted">
                  Provide your business name and choose a unique workspace URL.
                </p>
              </div>
            </div>

            <form onSubmit={apply} className="mt-5 space-y-5">
              {/* Business Name */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="tenant-name"
                  className="text-xs font-semibold text-foreground flex items-center justify-between"
                >
                  <span>
                    Business or Brand Name <span className="text-error">*</span>
                  </span>
                  <span className="text-[11px] font-normal text-muted">
                    {name.length}/100
                  </span>
                </Label>
                <Input
                  id="tenant-name"
                  required
                  minLength={2}
                  maxLength={100}
                  placeholder="e.g. Acme Fashion, Minh Long Ceramics"
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  disabled={busy}
                  className="h-10 text-sm bg-background/50 focus:bg-background"
                  autoFocus
                />
                <p className="text-[11px] text-muted">
                  The primary public name of your store or corporate organization.
                </p>
              </div>

              {/* Workspace Slug */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <Label
                    htmlFor="tenant-slug"
                    className="text-xs font-semibold text-foreground"
                  >
                    Workspace Slug <span className="text-error">*</span>
                  </Label>
                  {slugCustomized && (
                    <button
                      type="button"
                      onClick={handleResetSlug}
                      className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline"
                    >
                      <RotateCcw className="w-2.5 h-2.5" />
                      Reset to auto-generated
                    </button>
                  )}
                </div>

                <div className="relative">
                  <Input
                    id="tenant-slug"
                    required
                    minLength={2}
                    maxLength={100}
                    pattern="[a-z0-9]+(-[a-z0-9]+)*"
                    placeholder="e.g. acme-fashion"
                    value={slug}
                    onChange={(e) => handleSlugChange(e.target.value)}
                    disabled={busy}
                    className="h-10 font-mono text-xs pr-9 bg-background/50 focus:bg-background"
                  />
                  {isValidSlug && (
                    <div className="absolute right-3 top-1/2 -translate-y-1/2 text-success">
                      <CheckCircle2 className="w-4 h-4" />
                    </div>
                  )}
                </div>

                {/* Real-time URL preview */}
                <div className="flex items-center gap-2 rounded-xl border border-border/80 bg-background/70 px-3 py-2 text-xs text-muted font-mono">
                  <Globe className="w-3.5 h-3.5 text-primary shrink-0" />
                  <span className="truncate">
                    https://app.hexta.vn/tenants/
                    <span className="font-semibold text-foreground">
                      {slug || "your-slug"}
                    </span>
                  </span>
                </div>
                <p className="text-[11px] text-muted">
                  Used in browser bookmarks, URLs, and API endpoints. Lowercase letters, numbers, and hyphens only.
                </p>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-col-reverse sm:flex-row sm:items-center sm:justify-end gap-2.5 pt-3 border-t border-border/60">
                <Button
                  type="button"
                  variant="outline"
                  disabled={busy}
                  asChild
                  className="h-10 w-full sm:w-auto px-4 text-xs"
                >
                  <Link href="/tenants">Cancel</Link>
                </Button>

                <Button
                  type="submit"
                  disabled={busy || !name.trim() || !isValidSlug}
                  className="h-10 w-full sm:w-auto px-6 font-medium gap-2 shadow-sm text-xs"
                >
                  {busy ? (
                    <>
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                      <span>Submitting Request…</span>
                    </>
                  ) : (
                    <>
                      <Sparkles className="w-3.5 h-3.5" />
                      <span>Submit for Approval</span>
                      <ArrowRight className="w-3.5 h-3.5 ml-0.5" />
                    </>
                  )}
                </Button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
