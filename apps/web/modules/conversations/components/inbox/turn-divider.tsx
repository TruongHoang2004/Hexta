"use client";

/** History entry with centered text marking where one turn ended and the next began (or that the turn closed). */
export function TurnDivider({ children }: { children: string }) {
  return (
    <li className="flex items-center gap-3 text-xs text-muted">
      <span aria-hidden="true" className="h-px flex-1 bg-border" />
      <span className="text-center">{children}</span>
      <span aria-hidden="true" className="h-px flex-1 bg-border" />
    </li>
  );
}
