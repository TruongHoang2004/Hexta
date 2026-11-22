import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../lib/utils";

const badgeVariants = cva(
  "inline-flex items-center whitespace-nowrap rounded-full border px-2 py-0.5 text-xs font-medium",
  {
    variants: {
      tone: {
        neutral: "border-border bg-surface text-foreground",
        solid: "border-foreground bg-foreground text-background",
        warning: "border-warning bg-warning/15 text-foreground",
        success: "border-success bg-success/15 text-foreground",
        danger: "border-error bg-error/10 text-foreground",
        target: "border-dashed border-warning bg-surface text-muted",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export type BadgeTone = NonNullable<VariantProps<typeof badgeVariants>["tone"]>;

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

/** Small status label; meaning is carried by its text, the tone only reinforces it. */
function Badge({ className, tone, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ tone }), className)} {...props} />;
}

export { Badge, badgeVariants };
