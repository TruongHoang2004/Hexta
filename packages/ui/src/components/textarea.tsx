import * as React from "react";

import { cn } from "../lib/utils";
import { fieldStyles } from "./input";

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement>;

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(({ className, ...props }, ref) => (
  <textarea ref={ref} className={cn(fieldStyles, "min-h-20", className)} {...props} />
));
Textarea.displayName = "Textarea";

export { Textarea };
