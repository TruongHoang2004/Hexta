import type { ReactNode } from "react";

export interface TabItem<T extends string> {
  value: T;
  label: string;
  badge?: ReactNode;
}

/** ARIA tablist with arrow-key navigation; the caller renders the active panel. */
export function Tabs<T extends string>({
  label,
  items,
  value,
  onChange,
  idPrefix,
}: {
  label: string;
  items: TabItem<T>[];
  value: T;
  onChange: (value: T) => void;
  idPrefix: string;
}) {
  function onKeyDown(event: React.KeyboardEvent<HTMLButtonElement>, index: number) {
    const delta = event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
    if (!delta) return;
    event.preventDefault();
    const next = items[(index + delta + items.length) % items.length];
    onChange(next.value);
    document.getElementById(`${idPrefix}-tab-${next.value}`)?.focus();
  }
  return (
    <div role="tablist" aria-label={label} className="flex flex-wrap gap-1 border-b border-border">
      {items.map((item, index) => {
        const active = item.value === value;
        return (
          <button
            key={item.value}
            id={`${idPrefix}-tab-${item.value}`}
            type="button"
            role="tab"
            aria-selected={active}
            aria-controls={`${idPrefix}-panel`}
            tabIndex={active ? 0 : -1}
            onClick={() => onChange(item.value)}
            onKeyDown={(event) => onKeyDown(event, index)}
            className={`-mb-px inline-flex min-h-11 items-center gap-2 border-b-2 px-3 text-sm ${active ? "border-foreground font-semibold" : "border-transparent text-muted hover:text-foreground"}`}
          >
            {item.label}
            {item.badge}
          </button>
        );
      })}
    </div>
  );
}
