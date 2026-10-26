"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import type { TenantMember } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { getOverview, type OverviewResponse } from "@/lib/services/overview";
import { creationRange, localDate } from "@/lib/services/overview_dates";
import { errorText } from "@/lib/services/sales";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  Field,
  Notice,
} from "./primitives";

const timeFormat = new Intl.DateTimeFormat("en-GB", {
  timeZone: "Asia/Ho_Chi_Minh",
  dateStyle: "medium",
  timeStyle: "short",
});

export function OwnerOverview() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const query = useSearchParams().toString();
  if (!tenant || !user) return <p role="status">Loading account…</p>;
  if (tenant.owner_id !== user.id)
    return (
      <div role="alert" className="rounded-xl border border-border p-5">
        Only the current tenant owner can view this overview.
      </div>
    );
  return (
    <OverviewScope
      key={`${tenant.id}:${query}`}
      tenantID={tenant.id}
      userID={user.id}
      query={query}
    />
  );
}

function OverviewScope({
  tenantID,
  userID,
  query,
}: {
  tenantID: string;
  userID: string;
  query: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const initial = new URLSearchParams(query);
  const [allTime, setAllTime] = useState(
    !initial.has("from") && !initial.has("to"),
  );
  const [start, setStart] = useState(localDate(initial.get("from")));
  const [end, setEnd] = useState(localDate(initial.get("to"), true));
  const [employee, setEmployee] = useState(initial.get("employee_id") || "");
  const [assignment, setAssignment] = useState(
    initial.get("assignment") || "all",
  );
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [memberError, setMemberError] = useState("");
  const [membersLoaded, setMembersLoaded] = useState(false);
  const [data, setData] = useState<OverviewResponse | null>(null);
  const [error, setError] = useState("");
  const [formError, setFormError] = useState("");
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let ignore = false;
    sdk.identity
      .getUsers(tenantID)
      .then((rows) => {
        if (!ignore) {
          setMembers(rows);
          setMembersLoaded(true);
        }
      })
      .catch((err: unknown) => {
        if (!ignore) setMemberError(errorText(err));
      });
    return () => {
      ignore = true;
    };
  }, [tenantID]);
  useEffect(() => {
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      setData(null);
      try {
        const response = await getOverview(tenantID, query, abort.signal);
        if (!abort.signal.aborted) setData(response);
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, query, revision]);
  function navigate(params: URLSearchParams) {
    router.push(`${pathname}${params.size ? `?${params}` : ""}`);
  }
  function apply(event: React.FormEvent) {
    event.preventDefault();
    setFormError("");
    if (employee && assignment === "unassigned") {
      setFormError("Clear the employee filter to show unassigned records.");
      return;
    }
    const params = new URLSearchParams();
    try {
      if (!allTime) {
        const range = creationRange(start, end);
        params.set("from", range.from);
        params.set("to", range.to);
      }
    } catch (err) {
      setFormError(errorText(err));
      return;
    }
    if (employee) params.set("employee_id", employee);
    if (assignment !== "all") params.set("assignment", assignment);
    navigate(params);
  }
  function label(id: string | null) {
    if (!id) return "Unassigned";
    if (id === userID) return "You";
    return !membersLoaded || members.some((member) => member.user_id === id)
      ? id
      : `${id} (unavailable member)`;
  }
  const base = `/tenants/${tenantID}`;
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-3xl font-semibold">Owner overview</h1>
          <p className="mt-2 max-w-3xl text-sm text-muted">
            Current state of orders and work created in the selected period.
            Order and work creation dates are independent; these counts do not
            measure activity during the period.
          </p>
        </div>
        <button
          className={secondaryClass}
          disabled={loading}
          onClick={() => setRevision((value) => value + 1)}
        >
          Refresh saved records
        </button>
      </div>
      <form
        onSubmit={apply}
        className="space-y-4 rounded-xl border border-border bg-surface p-5"
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field title="Creation period">
            <select
              className={inputClass}
              value={allTime ? "all" : "range"}
              onChange={(event) => setAllTime(event.target.value === "all")}
            >
              <option value="all">All time</option>
              <option value="range">Date range</option>
            </select>
          </Field>
          <Field title="Assignment">
            <select
              className={inputClass}
              value={assignment}
              onChange={(event) => setAssignment(event.target.value)}
            >
              <option value="all">All assignments</option>
              <option value="assigned">Assigned</option>
              <option value="unassigned">Unassigned</option>
            </select>
          </Field>
          {!allTime && (
            <>
              <Field title="Start date (inclusive)">
                <input
                  className={inputClass}
                  type="date"
                  required
                  value={start}
                  onChange={(event) => setStart(event.target.value)}
                />
              </Field>
              <Field title="End date (inclusive)">
                <input
                  className={inputClass}
                  type="date"
                  required
                  min={start}
                  value={end}
                  onChange={(event) => setEnd(event.target.value)}
                />
              </Field>
            </>
          )}
          <Field title="Assigned sale / current work responsibility">
            <select
              className={inputClass}
              value={employee}
              onChange={(event) => setEmployee(event.target.value)}
            >
              <option value="">All employees</option>
              {employee &&
                !members.some((member) => member.user_id === employee) && (
                  <option value={employee}>{employee}</option>
                )}
              {members.map((member) => (
                <option key={member.id} value={member.user_id}>
                  {member.user_id === userID ? "You" : member.user_id}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <p className="text-xs text-muted">
          Dates and times use Asia/Ho_Chi_Minh (UTC+07:00). The end date
          includes the whole day. Waiting work belongs to its current
          responsible employee, separately from its nominated recipient.
        </p>
        <Notice error={formError} />
        {memberError && (
          <p role="alert" className="text-sm">
            Employee options unavailable: {memberError}
          </p>
        )}
        <div className="flex flex-wrap gap-2">
          <button className={buttonClass} type="submit">
            Apply filters
          </button>
          <Link className={secondaryClass} href={pathname}>
            Reset to all time
          </Link>
        </div>
      </form>
      <Notice error={error} />
      {error && (
        <p className="text-sm text-muted">
          Overview unavailable. Check your access and filters, then refresh to
          retry.
        </p>
      )}
      {loading && <p role="status">Loading saved overview…</p>}
      {data && !loading && (
        <>
          <p className="text-sm text-muted">
            Showing{" "}
            {data.filters.from && data.filters.to
              ? `${timeFormat.format(new Date(data.filters.from))} to ${timeFormat.format(new Date(data.filters.to))} (exclusive)`
              : "All time"}{" "}
            · {data.filters.assignment} assignments
            {data.filters.employee_id
              ? ` · ${label(data.filters.employee_id)}`
              : " · all employees"}
            . Filters apply to both count groups and work rows.
          </p>
          <section aria-labelledby="order-counts" className="space-y-3">
            <h2 id="order-counts" className="text-xl font-semibold">
              Orders
            </h2>
            <div className="grid gap-3 sm:grid-cols-3">
              {[
                ["Total orders", data.orders.total],
                ["Draft", data.orders.draft],
                ["Confirmed", data.orders.confirmed],
              ].map(([title, count]) => (
                <Count
                  key={title}
                  title={String(title)}
                  value={Number(count)}
                />
              ))}
            </div>
            <dl className="flex flex-wrap gap-x-5 gap-y-2 text-sm">
              {(
                [
                  "processing",
                  "fulfilled",
                  "cancelled",
                  "returned",
                  "other",
                ] as const
              ).map((status) => (
                <div key={status} className="flex gap-2">
                  <dt className="capitalize">{status}</dt>
                  <dd className="font-semibold">{data.orders[status]}</dd>
                </div>
              ))}
            </dl>
          </section>
          <section aria-labelledby="work-counts" className="space-y-3">
            <h2 id="work-counts" className="text-xl font-semibold">
              Work
            </h2>
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              {[
                ["Total work", data.work.total],
                ["Active", data.work.active],
                ["Waiting acceptance", data.work.waiting],
                ["Completed", data.work.completed],
              ].map(([title, count]) => (
                <Count
                  key={title}
                  title={String(title)}
                  value={Number(count)}
                />
              ))}
            </div>
            <p className="text-sm text-muted">
              Unassigned:{" "}
              <strong className="text-foreground">
                {data.work.unassigned}
              </strong>{" "}
              (included in active work).
            </p>
          </section>
          <section
            aria-labelledby="work-rows"
            className="space-y-4 rounded-xl border border-border bg-surface p-5"
          >
            <h2 id="work-rows" className="text-xl font-semibold">
              Matching work
            </h2>
            {!data.work_items.items.length ? (
              <p className="text-sm text-muted">
                {data.work.total
                  ? "No work on this page. Choose an earlier page."
                  : "No work matches these filters. Orders may exist without work."}
              </p>
            ) : (
              <ul className="space-y-3">
                {data.work_items.items.map((row) => (
                  <li
                    key={row.id}
                    className="space-y-2 rounded-lg border border-border p-4 text-sm"
                  >
                    <div className="flex flex-wrap justify-between gap-2">
                      <Link
                        className="font-semibold text-primary underline"
                        href={`${base}/work?work_id=${encodeURIComponent(row.id)}`}
                      >
                        {row.workflow_name} · v{row.workflow_version_number}
                      </Link>
                      <span className="capitalize">{row.state}</span>
                    </div>
                    <p>
                      Step {row.current_step + 1}: {row.current_step_name}
                    </p>
                    <p className="break-all">
                      Current responsible: {label(row.responsible_id)}
                    </p>
                    {row.recipient_id && (
                      <p className="break-all">
                        Nominated recipient: {label(row.recipient_id)}
                      </p>
                    )}
                    <p>
                      <Link
                        className="text-primary underline"
                        href={`${base}/orders?order_id=${encodeURIComponent(row.order_id)}`}
                      >
                        Order {row.order_number || row.order_id}
                      </Link>
                    </p>
                    <p className="text-xs text-muted">
                      Created {timeFormat.format(new Date(row.created_at))} ·
                      Updated {timeFormat.format(new Date(row.updated_at))}
                    </p>
                  </li>
                ))}
              </ul>
            )}
            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border pt-3 text-sm">
              <span>
                {data.work_items.total} records · Page {data.work_items.page} of{" "}
                {Math.max(
                  1,
                  Math.ceil(data.work_items.total / data.work_items.size),
                )}
              </span>
              <div className="flex gap-2">
                {[-1, 1].map((delta) => (
                  <button
                    key={delta}
                    className={secondaryClass}
                    disabled={
                      delta < 0
                        ? data.work_items.page <= 1
                        : data.work_items.page * data.work_items.size >=
                          data.work_items.total
                    }
                    onClick={() => {
                      const params = new URLSearchParams(query);
                      params.set("page", String(data.work_items.page + delta));
                      navigate(params);
                    }}
                  >
                    {delta < 0 ? "Previous" : "Next"}
                  </button>
                ))}
              </div>
            </div>
          </section>
        </>
      )}
    </div>
  );
}
function Count({ title, value }: { title: string; value: number }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <p className="text-sm text-muted">{title}</p>
      <p className="mt-2 text-3xl font-semibold">{value}</p>
    </div>
  );
}
