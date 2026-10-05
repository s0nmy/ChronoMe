import assert from "node:assert/strict";
import { test } from "node:test";
import { attachProjectsToEntries, deriveEntryTitle } from "../src/features/entries/entryHelpers.ts";

// These helpers only read IDs and project references, so the fixtures omit unrelated fields.
test("project edits are reflected without mutating stored entries", () => {
  const original = { id: "p1", name: "Before", color: "#000000" };
  const updated = { ...original, name: "After", color: "#ffffff" };
  const entry = { id: "e1", projectId: original.id, project: original };
  const [result] = attachProjectsToEntries([entry], [updated]);
  assert.equal(result.project, updated);
  assert.equal(entry.project, original);
});

test("missing and unassigned projects do not keep stale project data", () => {
  const stale = { id: "deleted" };
  const entries = [
    { id: "e1", projectId: stale.id, project: stale },
    { id: "e2", projectId: null },
    { id: "e3" },
  ];
  assert.deepEqual(
    attachProjectsToEntries(entries, []).map((entry) => entry.project),
    [undefined, undefined, undefined],
  );
});

test("unchanged project references preserve entries and their order", () => {
  const project = { id: "p1" };
  const entries = [{ id: "e1", projectId: "p1", project }, { id: "e2" }];
  const result = attachProjectsToEntries(entries, [project]);
  assert.equal(result[0], entries[0]);
  assert.equal(result[1], entries[1]);
});

test("entry titles prefer trimmed notes, then project name, then the default", () => {
  assert.equal(deriveEntryTitle("  Review  ", "Project"), "Review");
  assert.equal(deriveEntryTitle("  ", "Project"), "Projectの作業");
  assert.equal(deriveEntryTitle(), "作業ログ");
});
