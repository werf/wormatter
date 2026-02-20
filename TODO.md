# Wormatter Issues & Fixes (Found in nelm)

Issues discovered by formatting the [nelm](https://github.com/werf/nelm) codebase and analyzing results.

---

## Critical

### 1. Variable reordering breaks initialization dependencies (`pkg/featgate/feat.go`)
**Issue**: `FeatGates = []*FeatGate{}` moved to end of `var` block, overwriting populated slice (filled by `NewFeatGate` calls) with empty one.
**Fix**: Don't sort var specs by name at all. Only group by exportability (`_` first, exported, unexported) and preserve original relative order within each group. Dependency analysis via AST is unreliable (can't detect indirect dependencies through function calls), and alphabetical sorting can break initialization order.

### 2. Comment loss (`ChartTSEntryPoints`) (`pkg/common/common.go`)
**Issue**: Comment `// ChartTSEntryPoints defines supported TypeScript/JavaScript entry points (in priority order).` lost during var merge. Variable `ChartTSEntryPoints` was moved to a merged block, but the comment was detached.
**Fix**: When extracting specs from `GenDecl` in `collectGenDecl` (both `token.CONST` and `token.VAR` cases), transfer `GenDecl.Decs.Start` (doc comments) to the first `ValueSpec.Decs.Start` before discarding the `GenDecl`. This ensures comments travel with the spec during merge/sort. For multi-spec blocks, only transfer to the first spec (inner comments are already on the spec).

### 2a. Verify inline comment attachment during const reordering
**Issue**: Wormatter reorders `const` specs. We need tests ensuring inline comments remain attached to the correct spec after any reordering.
**Evidence**: `internal/plan/release_info.go` shows const spec reordering; comments still look correct there, but this should be covered by tests to prevent regressions.
**Fix**: No code change needed — DST decorations on `ValueSpec` (including inline `End` comments) travel with the node during `sort.SliceStable`. Add regression tests to verify inline comments remain attached to the correct spec after reordering.

### 2b. Verify free-floating comment handling
**Issue**: Need to verify that free-floating comments (not clearly attached to any decl/spec/block) are preserved and remain correctly positioned when wormatter merges/reorders declarations.
**Fix**: Detect free-floating comments (section headers separated from code by a blank line) before reordering. In DST, these are identifiable by a trailing `"\n"` entry in `Decs.Start`. If any top-level declaration has such a detached comment, return an error for that file ("file has free-floating comments, cannot safely reorder declarations"). Add tests for this detection and for normal doc comments (no trailing `"\n"`) being preserved correctly.

---

## Major

### 3. Test table struct field reordering buries `name` field
**Issue**: `name` field moved to end of test structs (alphabetical sort), making table tests unreadable.
**Fix**: Only reorder struct fields in named type declarations (`type Foo struct{...}`). Change `reorderStructFields` to look for `*dst.TypeSpec` nodes and reorder the `*dst.StructType` inside them, instead of targeting all `*dst.StructType` nodes indiscriminately. This skips anonymous structs (table tests, inline struct fields, composite literal types).

### 4. JSON-serialized struct field reordering changes wire format
**Issue**: Struct field order changes JSON output order in Go; wormatter reorders json-tagged fields in exported types (e.g. `internal/plan/operation.go`, `internal/plan/plan.go`).
**Fix**: If any field in a struct has an encoding-related struct tag (`json`, `yaml`, `xml`, `toml`, `protobuf`), skip field reordering for that entire struct (preserve original field order). Don't mix sorted and unsorted fields — it creates confusing output. Non-encoding tags (`validate`, `db`, etc.) don't trigger the skip.

---

## Minor

### 5. Spurious blank lines in structs
**Issue**: Blank lines inserted between fields in both public and private structs (e.g., `NoActivityTimeout` / `Ownership` in public structs; `logStore` / `maxLogEventTableWidth` in private `tablesBuilder` struct in `internal/track/progress_tables.go`; `discoveryClient` / `dynamicClient` in private `KubeClient` struct in `internal/kube/client_kube.go`).
**Fix**: The grouping logic in `assembleFieldList` is correct (only sets `EmptyLine` at group boundaries), but the loops preserve stale `EmptyLine` decorations from the original source via a `if f.Decs.Before != dst.EmptyLine` guard. Fix: unconditionally set `Decs.Before = dst.NewLine` for all fields within a group, then set `EmptyLine` only on the first field of each new group.

### 6. Ordered typed string constants reordered
**Issue**: Typed string constant blocks that are intentionally ordered get reordered alphabetically (e.g. stage ordering in `pkg/common/common.go`, and previously observed `ReleaseType` constants).
**Fix**: Create a const-specific sort function (separate from vars). Sort const specs by: exportability → type name → name **only for untyped consts** (empty type). Typed consts (`Stage`, `ReleaseType`, etc.) preserve their original relative order within the same type group — `sort.SliceStable` handles this naturally by returning `false` for same-type comparisons. This keeps intentional orderings (pipeline stages, priorities) intact while still alphabetizing untyped consts.

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
