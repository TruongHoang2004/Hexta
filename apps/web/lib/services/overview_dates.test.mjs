import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import ts from "typescript";

// Compile the small pure helper so this test also runs on the CI Node 20 baseline.
const source = readFileSync(
  new URL("./overview_dates.ts", import.meta.url),
  "utf8",
);
const { outputText } = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2020,
  },
});
const { creationRange, localDate } = await import(
  `data:text/javascript;base64,${Buffer.from(outputText).toString("base64")}`
);

test("inclusive local dates become offset-aware half-open creation bounds", () => {
  const range = creationRange("2026-10-19", "2026-10-25");
  assert.equal(new Date(range.from).toISOString(), "2026-10-18T17:00:00.000Z");
  assert.equal(new Date(range.to).toISOString(), "2026-10-25T17:00:00.000Z");
  assert.equal(localDate("2026-10-18T17:00:00Z"), "2026-10-19");
  assert.equal(localDate("2026-10-25T17:00:00Z", true), "2026-10-25");
  assert.match(new URLSearchParams(range).toString(), /%2B07%3A00/);
});

test("same-day, leap-day and year-end ranges include the final whole day", () => {
  for (const [date, next] of [
    ["2026-10-23", "2026-10-24"],
    ["2028-02-29", "2028-03-01"],
    ["2026-12-31", "2027-01-01"],
  ]) {
    const range = creationRange(date, date);
    assert.equal(range.to, `${next}T00:00:00+07:00`);
    assert.equal(
      new Date(range.to) - new Date(range.from),
      24 * 60 * 60 * 1000,
    );
  }
});

test("invalid, missing and reversed date ranges fail before navigation", () => {
  for (const [start, end] of [
    ["", "2026-10-23"],
    ["2026-02-29", "2026-03-01"],
    ["2026-10-24", "2026-10-23"],
    ["2026-13-01", "2026-13-02"],
  ]) {
    assert.throws(() => creationRange(start, end));
  }
  assert.equal(localDate("invalid"), "");
});
