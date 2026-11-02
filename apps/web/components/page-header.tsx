"use client";

import Link from "next/link";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { Button } from "@hexta/ui";

export interface PageHeaderProps {
  /** Optional target URL to navigate back to */
  backHref?: string;
  /** Label for the back link or breadcrumb, e.g. "Workspaces" */
  backLabel?: string;
  /** Primary page heading */
  title: string;
  /** Secondary subtitle or helper text */
  description?: string;
  /** Optional badge or status tag placed next to title */
  badge?: React.ReactNode;
  /** Optional right-aligned action buttons or tools */
  actions?: React.ReactNode;
  /** Additional container styling */
  className?: string;
}

export function PageHeader({
  backHref,
  backLabel,
  title,
  description,
  badge,
  actions,
  className = "",
}: PageHeaderProps) {
  return (
    <div className={`space-y-3 border-b border-border/60 pb-6 ${className}`}>
      {/* Compact Breadcrumb Bar (Zero vertical bloat) */}
      {backHref && (
        <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 text-xs text-muted">
          <Link
            href={backHref}
            className="group inline-flex items-center gap-1.5 font-medium hover:text-foreground transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5 transition-transform group-hover:-translate-x-0.5" />
            <span>{backLabel || "Back"}</span>
          </Link>
          <ChevronRight className="w-3 h-3 text-muted/50 shrink-0" />
          <span className="font-semibold text-foreground truncate max-w-[220px] sm:max-w-md">
            {title}
          </span>
        </nav>
      )}

      {/* Main Title Row */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-3">
          {backHref && !backLabel && (
            <Button
              variant="outline"
              size="icon"
              asChild
              className="h-9 w-9 shrink-0 rounded-xl border-border/80 text-muted hover:text-foreground shadow-2xs mt-0.5"
            >
              <Link href={backHref} aria-label="Go back">
                <ArrowLeft className="w-4 h-4" />
              </Link>
            </Button>
          )}

          <div>
            <div className="flex flex-wrap items-center gap-2.5">
              <h1 className="text-2.5xl sm:text-3xl font-bold tracking-tight text-foreground">
                {title}
              </h1>
              {badge}
            </div>
            {description && (
              <p className="mt-1 text-sm text-muted leading-relaxed max-w-2xl">
                {description}
              </p>
            )}
          </div>
        </div>

        {actions && <div className="flex items-center gap-2.5 shrink-0">{actions}</div>}
      </div>
    </div>
  );
}
