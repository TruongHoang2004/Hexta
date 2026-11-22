import Link from "next/link";

/** A sidebar link; marks the current page for assistive technology. */
export function NavLink({ href, active, children }: { href: string; active: boolean; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={`flex min-h-9 items-center rounded-lg px-3 text-sm ${active ? "bg-primary/10 font-semibold text-foreground" : "text-muted hover:bg-background hover:text-foreground"}`}
    >
      {children}
    </Link>
  );
}
