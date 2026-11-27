"use client";

import type { TenantMember } from "@hexta/sdk";
import { Field, Picker } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OverviewFilter } from "../types";
import { overviewPeriods } from "../utils";

interface Props {
  filter: OverviewFilter;
  members: TenantMember[];
  person: (userID: string | null | undefined) => string;
  onChange: (filter: OverviewFilter) => void;
}

/** Period preset and employee pickers. */
export function OverviewFilters({ filter, members, person, onChange }: Props) {
  const t = useT();
  const periodOptions = overviewPeriods.map((period) => ({ value: period, label: t(`overview.periods.${period}`) }));
  const employeeOptions = [
    { value: "", label: t("overview.filters.allEmployees") },
    ...members.map((member) => ({ value: member.user_id, label: person(member.user_id) })),
  ];
  return (
    <div className="space-y-1">
      <div className="flex flex-wrap gap-3">
        <div className="w-44">
          <Field title={t("overview.filters.period")}>
            <Picker
              options={periodOptions}
              value={filter.period}
              onChange={(value) => onChange({ ...filter, period: overviewPeriods.find((p) => p === value) ?? filter.period })}
            />
          </Field>
        </div>
        <div className="w-60">
          <Field title={t("overview.filters.employee")}>
            <Picker
              options={employeeOptions}
              value={filter.employeeID}
              aria-describedby="overview-employee-hint"
              onChange={(value) => onChange({ ...filter, employeeID: value })}
            />
          </Field>
        </div>
      </div>
      <p id="overview-employee-hint" className="text-xs text-muted">
        {t("overview.filters.employeeHint")}
      </p>
    </div>
  );
}
