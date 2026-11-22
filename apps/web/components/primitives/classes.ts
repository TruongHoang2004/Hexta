import { buttonVariants, cn, fieldStyles } from "@hexta/ui";

// Class strings for places that style a native element or a Link directly; they come from the
// @hexta/ui variants so plain elements and shared components look the same.
export const inputClass = cn(fieldStyles, "min-h-10");
export const buttonClass = buttonVariants();
export const secondaryClass = buttonVariants({ variant: "outline" });
