"use client";

import * as React from "react";
import { Command } from "cmdk";
import { Check, ChevronsUpDown, Search } from "lucide-react";

import { cn } from "../lib/utils";
import { fieldStyles } from "./input";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";

export interface ComboboxOption {
  /** "" is a valid value, e.g. for "Unassigned" or "Choose later". */
  value: string;
  label: string;
  /** Secondary text shown under the label and matched by search. */
  description?: string;
  disabled?: boolean;
}

export interface ComboboxProps {
  options: ComboboxOption[];
  value: string;
  onChange: (value: string) => void;
  /** Trigger text when no option matches `value`. */
  placeholder: string;
  searchPlaceholder: string;
  emptyText: string;
  disabled?: boolean;
  /** Blocks form submission while the value is "", like a native required select. */
  required?: boolean;
  name?: string;
  id?: string;
  className?: string;
  "aria-label"?: string;
  "aria-describedby"?: string;
}

/**
 * Searchable single-choice picker (Popover + cmdk) replacing native selects. The trigger is a
 * button, so a wrapping <label> names it; arrows, Enter and Escape work inside the list.
 */
export function Combobox({
  options,
  value,
  onChange,
  placeholder,
  searchPlaceholder,
  emptyText,
  disabled,
  required,
  name,
  id,
  className,
  ...aria
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);
  const listID = React.useId();
  const selected = options.find((option) => option.value === value);
  const validity = React.useRef<HTMLInputElement>(null);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <span className="relative block">
        <PopoverTrigger asChild>
          <button
            id={id}
            type="button"
            role="combobox"
            aria-expanded={open}
            aria-controls={listID}
            aria-required={required || undefined}
            disabled={disabled}
            className={cn(fieldStyles, "flex h-10 items-center justify-between gap-2 text-left font-normal", className)}
            {...aria}
          >
            <span className={cn("truncate", !selected && "text-muted")}>{selected ? selected.label : placeholder}</span>
            <ChevronsUpDown className="h-4 w-4 shrink-0 text-muted" aria-hidden="true" />
          </button>
        </PopoverTrigger>
        {(required || name) && (
          // Carries the value into native form validation/submission; the bubble points at the trigger area.
          <input
            ref={validity}
            tabIndex={-1}
            aria-hidden="true"
            name={name}
            required={required}
            value={value}
            onChange={() => undefined}
            onInvalid={() => setOpen(true)}
            className="pointer-events-none absolute inset-x-0 bottom-0 h-px w-full opacity-0"
          />
        )}
      </span>
      <PopoverContent className="w-(--radix-popover-trigger-width) min-w-64 p-0">
        <Command
          id={listID}
          // Match label and description, not the raw value (IDs).
          filter={(itemValue, search) => (itemValue.toLowerCase().includes(search.trim().toLowerCase()) ? 1 : 0)}
        >
          <div className="flex items-center gap-2 border-b border-border px-3">
            <Search className="h-4 w-4 shrink-0 text-muted" aria-hidden="true" />
            <Command.Input
              placeholder={searchPlaceholder}
              className="h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted"
            />
          </div>
          <Command.List className="max-h-72 overflow-y-auto p-1">
            <Command.Empty className="px-3 py-6 text-center text-sm text-muted">{emptyText}</Command.Empty>
            {options.map((option) => (
              <Command.Item
                key={option.value || "__empty"}
                value={`${option.label} ${option.description ?? ""} ${option.value}`}
                disabled={option.disabled}
                onSelect={() => {
                  onChange(option.value);
                  setOpen(false);
                }}
                className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-2 text-sm data-[disabled=true]:cursor-not-allowed data-[selected=true]:bg-background data-[disabled=true]:opacity-50"
              >
                <Check
                  className={cn("mt-0.5 h-4 w-4 shrink-0", option.value === value ? "opacity-100" : "opacity-0")}
                  aria-hidden="true"
                />
                <span className="min-w-0">
                  <span className="block truncate">{option.label}</span>
                  {option.description && <span className="block truncate text-xs text-muted">{option.description}</span>}
                </span>
              </Command.Item>
            ))}
          </Command.List>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
