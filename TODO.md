# Wormatter Issues & Fixes (Found in nelm)

Issues discovered by formatting the [nelm](https://github.com/werf/nelm) codebase and analyzing results.

---

## Critical

### ~~1. Variable reordering breaks initialization dependencies~~ → Fixed
**Fix applied**: Var specs are now grouped by exportability only (`_` first, exported, unexported) and preserve original relative order within each group. `sortVarSpecsByExportability` replaces `sortSpecsByExportabilityThenName` for vars.

### ~~2. Comment loss (`ChartTSEntryPoints`)~~ → Fixed
**Fix applied**: `transferGenDeclDocToFirstSpec` transfers `GenDecl.Decs.Start` to the first `ValueSpec.Decs.Start` in `collectGenDecl` for both `token.CONST` and `token.VAR` cases.

### ~~2a. Verify inline comment attachment during const reordering~~ → Fixed
**Fix applied**: Added regression tests. Also fixed end-comment transfer: `GenDecl.Decs.End` (inline comments on single-spec declarations) is now transferred to the spec's `Decs.End`. DST limitation: the last spec's `End` comment in a block is moved to `Start` via `fixLastSpecEndComment` to prevent misplacement.

### ~~2b. Verify free-floating comment handling~~ → Fixed
**Fix applied**: `checkFreeFloatingComments` detects free-floating comments (trailing `"\n"` in `Decs.Start`) on var/const declarations before reordering and returns an error.

---

## Major

### ~~3. Test table struct field reordering buries `name` field~~ → Fixed
**Fix applied**: `reorderStructFields` now inspects `*dst.TypeSpec` nodes and reorders the `*dst.StructType` inside them, skipping anonymous structs (table tests, inline struct fields, composite literal types).

### ~~4. JSON-serialized struct field reordering changes wire format~~ → Fixed
**Fix applied**: `hasEncodingTags` checks if any field has `json`, `yaml`, `xml`, `toml`, or `protobuf` struct tags. If so, `reorderStructFields` skips that struct entirely.

---

## Minor

### ~~5. Spurious blank lines in structs~~ → Fixed
**Fix applied**: `assembleFieldList` now unconditionally sets `Decs.Before = dst.NewLine` and `Decs.After = dst.None` for all fields, then sets `EmptyLine` only on the first field of each new group boundary.

### ~~6. Ordered typed string constants reordered~~ → Fixed
**Fix applied**: `sortConstSpecsByExportabilityThenName` sorts by exportability → type name → name only for untyped consts. Typed consts preserve their original relative order within the same type group via `sort.SliceStable`.

### 7. Table test cases reordered in slice literals
**Issue**: Elements in table-driven test case slices appear to be reordered (e.g. `internal/plan/plan_build_test.go`, where `{ name: `...`, input: ..., expect: ... }` cases moved around). This is noisy at best and can be semantically risky if test cases are order-dependent.
**Fix**: Duplicate of #3 + #8. The code does not reorder slice elements — the perceived reordering is caused by struct field reordering *within* each element (anonymous struct fields sorted by #3, keyed literal fields reordered by #8). Fixing those issues resolves this.

### ~~8. Keyed composite literals reordered (struct literals)~~ → WontFix
**Issue**: Keyed elements inside struct literals are being reordered to match struct definition order. This can theoretically change side-effect order (Go evaluates element expressions in source order).
**Decision**: Keep reordering. Side-effect-dependent composite literal values are rare and a code smell. The reordering is intentional behavior.

### ~~9. Map literal entries reordered~~ → Misdiagnosed / WontFix
**Issue**: Map literal entries appeared to be reordered in `internal/resource/sensitive_test.go`.
**Analysis**: The code does NOT reorder map entries — `resolveSortedFieldOrder` returns `nil` for `*dst.MapType`, so `reorderCompositeLitFields` is never called for maps. The perceived reordering was caused by other changes (struct field/declaration reordering) shifting surrounding code in the diff.

---

## WontFix / Working as Intended

- Function signature collapsing: `runFailurePlan` (507 chars) stays single-line. User preference is "don't touch".
- Test method/function reordering: Test methods remain sorted alphabetically. User preference is "do nothing".
- Embedded struct reordering: `*Options` structs continue sorting embedded fields alphabetically. User preference is "Sort Alphabetically".
- Const block merging: The formatter merges const blocks and groups by type within the merged block. While this changes the source layout (types detached from consts), it's consistent with the "one const block per file" style. The grouping by type inside the block is sufficient.
- Reordering `init()` relative to package-level initializers (was #2c): Safe per the Go spec — all package-level variables are initialized before any `init()` runs, regardless of source position. The current implementation preserves the relative order of multiple `init()` functions within the same file (collected and emitted in source order via `collector.go`).
