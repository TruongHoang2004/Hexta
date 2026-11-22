/** One row of a pass/fail checklist (e.g. order confirmation conditions). */
export function ChecklistItem({
  passed,
  label,
  hint,
  passedText,
  failedText,
}: {
  passed: boolean;
  label: string;
  hint?: string;
  passedText: string;
  failedText: string;
}) {
  return (
    <li className="flex items-start gap-2.5 border-b border-border py-2 text-sm last:border-b-0">
      <span
        aria-hidden="true"
        className={`mt-0.5 inline-flex h-5 w-5 flex-none items-center justify-center rounded-full text-xs font-bold ${passed ? "bg-success text-white" : "border-2 border-error text-error"}`}
      >
        {passed ? "✓" : "!"}
      </span>
      <span>
        <span className="sr-only">{passed ? passedText : failedText}: </span>
        <span className={passed ? "" : "font-semibold"}>{label}</span>
        {hint && <span className="block text-muted">{hint}</span>}
      </span>
    </li>
  );
}
