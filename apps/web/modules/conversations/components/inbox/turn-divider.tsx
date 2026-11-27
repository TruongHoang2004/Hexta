"use client";

/**
 * History entry with centered text marking where a session (turn) starts or that the current one closed. With an
 * `id` it is a jump target for the session list.
 */
export function TurnDivider({ id, children }: { id?: string; children: string }) {
  return (
    <li
      id={id}
      tabIndex={id ? -1 : undefined}
      className="flex scroll-mt-2 items-center gap-3 text-xs text-muted focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary"
    >
      <span aria-hidden="true" className="h-px flex-1 bg-border" />
      <span className="text-center">{children}</span>
      <span aria-hidden="true" className="h-px flex-1 bg-border" />
    </li>
  );
}
