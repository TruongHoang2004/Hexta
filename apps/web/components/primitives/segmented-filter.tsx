export interface FilterOption<T extends string> {
  value: T;
  label: string;
  count?: number;
}

/** Single-choice filter rendered as toggle buttons. */
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
    <div role="group" aria-label={label} className="flex flex-wrap gap-1.5">
      {options.map((option) => {
        const active = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(option.value)}
            className={`min-h-9 rounded-full border px-3 text-sm ${active ? "border-foreground bg-foreground text-background" : "border-border bg-surface hover:bg-background"}`}
          >
            {option.label}
            {option.count !== undefined && <span className="ml-1.5 opacity-80">{option.count}</span>}
          </button>
        );
      })}
    </div>
  );
}
