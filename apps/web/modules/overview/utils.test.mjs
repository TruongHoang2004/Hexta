import { test } from "node:test";
import assert from "node:assert/strict";
import { groupByWorkflow, isEmptySelection, periodRange, waitedFor } from "./utils.ts";

test("periodRange covers whole local days ending today, half-open", () => {
  const now = new Date(2026, 10, 26, 15, 30);
  assert.deepEqual(periodRange("all", now), {});
  const today = periodRange("today", now);
  assert.equal(new Date(today.from).getTime(), new Date(2026, 10, 26).getTime());
  assert.equal(new Date(today.to).getTime(), new Date(2026, 10, 27).getTime());
  const week = periodRange("7d", now);
  assert.equal(new Date(week.from).getTime(), new Date(2026, 10, 20).getTime());
  assert.equal(new Date(week.to).getTime(), new Date(2026, 10, 27).getTime());
  const month = periodRange("30d", now);
  assert.equal(new Date(month.from).getTime(), new Date(2026, 9, 28).getTime());
});

test("waitedFor picks the largest whole unit", () => {
  const now = new Date("2026-11-26T12:00:00Z");
  assert.deepEqual(waitedFor("2026-11-26T11:59:50Z", now), { unit: "minutes", value: 1 });
  assert.deepEqual(waitedFor("2026-11-26T07:00:00Z", now), { unit: "hours", value: 5 });
  assert.deepEqual(waitedFor("2026-11-23T11:00:00Z", now), { unit: "days", value: 3 });
});

test("isEmptySelection distinguishes an empty workspace from an empty period", () => {
  const data = (hasActivity, orders, work, stuck) => ({
    has_activity: hasActivity,
    orders: { total: orders },
    work: { total: work },
    stuck: { total: stuck },
  });
  assert.equal(isEmptySelection(data(false, 0, 0, 0)), false);
  assert.equal(isEmptySelection(data(true, 0, 0, 0)), true);
  assert.equal(isEmptySelection(data(true, 1, 0, 0)), false);
});

test("groupByWorkflow keeps pinned versions apart", () => {
  const step = (version, ordinal) => ({
    workflow_version_id: version,
    workflow_name: "Delivery",
    workflow_version_number: version === "v1" ? 1 : 2,
    step_ordinal: ordinal,
    step_name: `S${ordinal}`,
    active: 1,
    waiting: 0,
    completed: 0,
    unassigned: 0,
  });
  const groups = groupByWorkflow([step("v1", 0), step("v1", 1), step("v2", 0)]);
  assert.equal(groups.length, 2);
  assert.deepEqual(groups.map((g) => [g.version, g.steps.length]), [[1, 2], [2, 1]]);
});
