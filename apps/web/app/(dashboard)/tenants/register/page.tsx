"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { onboarding } from "@/lib/services/onboarding";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
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
  const t = useT();
  const errorText = useErrorText();
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
      toast.error(t("registerWorkspace.nameTooShort"));
      return;
    }

    if (!trimmedSlug || !/^[a-z0-9]+(-[a-z0-9]+)*$/.test(trimmedSlug)) {
      toast.error(t("registerWorkspace.slugInvalid"));
      return;
    }

    setBusy(true);
    try {
      await onboarding.apply(trimmedName, trimmedSlug);
      toast.success(t("registerWorkspace.submitted"));
      router.push("/tenants");
    } catch (err) {
      toast.error(errorText(err));
      setBusy(false);
    }
  }

  const [privacyBefore, privacyAfter] = t("registerWorkspace.privacy").split("{link}");
  const isValidSlug = slug.length >= 2 && /^[a-z0-9]+(-[a-z0-9]+)*$/.test(slug);

  return (
    <div className="space-y-6 pb-12">
      {/* Unified Compact PageHeader with integrated Breadcrumb/Back link */}
      <PageHeader
        backHref="/tenants"
        backLabel={t("registerWorkspace.back")}
        title={t("registerWorkspace.title")}
        description={t("registerWorkspace.description")}
        badge={
          <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
            <Sparkles className="w-3 h-3" /> {t("registerWorkspace.badge")}
          </span>
        }
      />

      <div className="grid gap-8 lg:grid-cols-12 items-start pt-1">
        {/* Left Side: Value Proposition & Onboarding Guide */}
        <div className="lg:col-span-5 space-y-5">
          <div>
            <h2 className="text-sm font-semibold text-foreground">
              {t("registerWorkspace.includedTitle")}
            </h2>
            <p className="mt-0.5 text-xs text-muted leading-relaxed">
              {t("registerWorkspace.includedBody")}
            </p>
          </div>

          <div className="space-y-3">
            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <MessageSquare className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  {t("registerWorkspace.inboxTitle")}
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  {t("registerWorkspace.inboxBody")}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Bot className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  {t("registerWorkspace.aiTitle")}
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  {t("registerWorkspace.aiBody")}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/60 p-3.5 transition-colors">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Workflow className="w-4.5 h-4.5" />
              </div>
              <div>
                <h3 className="text-xs font-semibold text-foreground">
                  {t("registerWorkspace.workflowTitle")}
                </h3>
                <p className="mt-0.5 text-[11px] text-muted leading-relaxed">
                  {t("registerWorkspace.workflowBody")}
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-start gap-3 rounded-2xl border border-border/60 bg-muted/5 p-3.5 text-xs text-muted">
            <ShieldCheck className="w-4 h-4 text-primary shrink-0 mt-0.5" />
            <div className="leading-relaxed text-[11px]">
              <span className="font-semibold text-foreground">{t("registerWorkspace.reviewTitle")}</span>
              {t("registerWorkspace.reviewBody")}
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
                  {t("registerWorkspace.formTitle")}
                </h2>
                <p className="text-xs text-muted">
                  {t("registerWorkspace.formBody")}
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
                    {t("registerWorkspace.name")} <span className="text-error">*</span>
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
                  placeholder={t("registerWorkspace.namePlaceholder")}
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  disabled={busy}
                  className="h-10 text-sm bg-background/50 focus:bg-background"
                  autoFocus
                />
                <p className="text-[11px] text-muted">
                  {t("registerWorkspace.nameHelp")}
                </p>
              </div>

              {/* Workspace Slug */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <Label
                    htmlFor="tenant-slug"
                    className="text-xs font-semibold text-foreground"
                  >
                    {t("registerWorkspace.slug")} <span className="text-error">*</span>
                  </Label>
                  {slugCustomized && (
                    <button
                      type="button"
                      onClick={handleResetSlug}
                      className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline"
                    >
                      <RotateCcw className="w-2.5 h-2.5" />
                      {t("registerWorkspace.slugReset")}
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
                    placeholder={t("registerWorkspace.slugPlaceholder")}
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
                    {t("registerWorkspace.slugPreview", { slug: slug || t("registerWorkspace.slugFallback") })}
                  </span>
                </div>
                <p className="text-[11px] text-muted">
                  {t("registerWorkspace.slugHelp")}
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
                  <Link href="/tenants">{t("registerWorkspace.cancel")}</Link>
                </Button>

                <Button
                  type="submit"
                  disabled={busy || !name.trim() || !isValidSlug}
                  className="h-10 w-full sm:w-auto px-6 font-medium gap-2 shadow-sm text-xs"
                >
                  {busy ? (
                    <>
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                      <span>{t("registerWorkspace.submitting")}</span>
                    </>
                  ) : (
                    <>
                      <Sparkles className="w-3.5 h-3.5" />
                      <span>{t("registerWorkspace.submit")}</span>
                      <ArrowRight className="w-3.5 h-3.5 ml-0.5" />
                    </>
                  )}
                </Button>
              </div>

              <p className="text-center text-[11px] text-muted">
                {privacyBefore}
                <Link href="/privacy" className="text-primary hover:underline underline-offset-2">
                  {t("registerWorkspace.privacyLink")}
                </Link>
                {privacyAfter}
              </p>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
