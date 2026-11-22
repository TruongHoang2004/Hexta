import * as React from "react"

import { cn } from "../lib/utils"

/** Shared field styling for Input, Textarea and the Combobox trigger. */
export const fieldStyles =
  "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground shadow-sm transition-colors placeholder:text-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:cursor-not-allowed disabled:opacity-50"

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          fieldStyles,
          "flex h-10 file:border-0 file:bg-transparent file:text-sm file:font-medium",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }
