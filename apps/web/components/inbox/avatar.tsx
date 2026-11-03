export function Avatar({ initials }: { initials: string }) {
  return (
    <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-violet-100 text-xs font-semibold text-violet-700">
      {initials}
    </span>
  );
}
