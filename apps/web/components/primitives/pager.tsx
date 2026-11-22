import { useT } from "@/lib/i18n/provider";
import { secondaryClass } from "./classes";

/** Previous/next pager for API pages of 20 rows. */
export function Pager({
  page,
  total,
  onChange,
  disabled,
}: {
  page: number;
  total: number;
  onChange: (page: number) => void;
  disabled?: boolean;
}) {
  const t = useT();
  return (
    <div className="flex items-center justify-between gap-3 border-t border-border pt-3 text-sm">
      <span>{t("common.pager", { total, page, pages: Math.max(1, Math.ceil(total / 20)) })}</span>
      <div className="flex gap-2">
        <button
          type="button"
          className={secondaryClass}
          disabled={disabled || page <= 1}
          onClick={() => onChange(page - 1)}
        >
          {t("common.previous")}
        </button>
        <button
          type="button"
          className={secondaryClass}
          disabled={disabled || page * 20 >= total}
          onClick={() => onChange(page + 1)}
        >
          {t("common.next")}
        </button>
      </div>
    </div>
  );
}
