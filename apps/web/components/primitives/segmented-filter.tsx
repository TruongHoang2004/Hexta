"use client";

import { ToggleGroup, ToggleGroupItem } from "@hexta/ui";

export interface FilterOption<T extends string> {
  value: T;
  label: string;
  count?: number;
}

/** Single-choice filter rendered as toggles; one option always stays selected. */
export function SegmentedFilter<T extends string>({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: FilterOption<T>[];
  value: T;
  onChange: (value: T) => void;
}) {
  return (
    <ToggleGroup
      type="single"
      aria-label={label}
      value={value}
      onValueChange={(next) => {
        const option = options.find((row) => row.value === next);
        if (option) onChange(option.value);
      }}
    >
      {options.map((option) => (
        <ToggleGroupItem key={option.value} value={option.value}>
          {option.label}
          {option.count !== undefined && <span className="ml-1.5 opacity-80">{option.count}</span>}
        </ToggleGroupItem>
      ))}
    </ToggleGroup>
  );
}
