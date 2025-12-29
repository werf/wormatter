# Wormatter

A DST-based Go source code formatter. Highly opinionated, but very comprehensive. Gofumpt and gci built-in.

## Installation

Download the latest binary from [GitHub Releases](https://github.com/werf/wormatter/releases).

## Usage

```bash
wormatter <file.go|directory>
```

Formats Go files in place. Recursively processes directories.

## Building

```bash
task build
```

## Formatting Rules

### Declaration Order (top to bottom)

1. **Imports** — unchanged
2. **init functions** — preserved in original order
3. **Constants** — merged into single `const()` block
4. **Variables** — merged into single `var()` block
5. **Types** — grouped by category, each followed by constructors and methods
6. **Standalone functions** — sorted by exportability
7. **main function** — always last

### Const/Var Block Rules

#### Grouping (separated by empty lines)
1. Blank identifier (`var _ Interface = ...`)
2. Public (uppercase first letter)
3. Private (lowercase first letter)

#### Within each group
- Sorted alphabetically by first name
- No empty lines between elements

#### Block format
- Single declaration: `const X = 1`
- Multiple declarations: `const ( ... )` with parentheses

### Type Grouping Order

Types are grouped by category (in this order):
1. Simple types (aliases like `type MyString string`, function types)
2. Function interfaces (interfaces with exactly one method)
3. Non-function interfaces (interfaces with 0 or 2+ methods)
4. Structs

Types within each category preserve their original order.

### Type-Associated Declarations

After each type definition:
1. **Constructors** — functions starting with `New` or `new` that return the type
2. **Methods** — functions with receiver of that type

#### Constructor Matching

A function is a constructor for type `T` if:
- Name starts with `New` (exported) or `new` (unexported)
- Returns `T`, `*T`, `(T, error)`, `(*T, error)`, etc.
- Name after `New`/`new` matches `T` case-insensitively, or starts with `T` followed by non-lowercase char
  - `NewFoo` ✓ matches `Foo`
  - `newFoo` ✓ matches `foo` (case-insensitive)
  - `NewFooWithOptions` ✓ matches `Foo`
  - `newFooWithOptions` ✓ matches `foo`
  - `NewFoobar` ✗ does NOT match `Foo` (would match `Foobar`)

#### Constructor/Method Sorting
- Constructors: alphabetically by name
- Methods: exported first, then unexported; within each group sorted by architectural layer

#### Standalone Functions Sorting
- Exported first, then unexported
- Within each group: sorted by architectural layer (high-level first, utilities last)

### Architectural Layer

Layer is determined by call depth to other local functions:
- Layer 0: functions that call no other local functions (leaves/utilities)
- Layer 1: functions that only call layer 0 functions
- Layer N: functions that call layer N-1 or lower
- Cyclic calls: functions in a cycle share the same layer

Higher layers appear first (orchestrators before utilities).

### Struct Field Ordering

Fields grouped into three blocks (separated by empty lines):
1. **Embedded** — fields without names, sorted alphabetically by type name
2. **Public** — uppercase names, sorted alphabetically
3. **Private** — lowercase names, sorted alphabetically

### Struct Literal Ordering

When instantiating structs with named fields:
```go
&Config{Timeout: 30, Verbose: true, debug: false}
```
Fields are reordered to match struct definition order (embedded → public → private).
No empty lines between fields in literals.

### Function Body Rules

#### One-line functions
- Empty body stays one line: `func foo() {}`
- Non-empty body expands to multiple lines

#### Return statements
- Empty line before `return` if there's code before it in the same block
- No empty line if `return` is the first/only statement

### Spacing Rules

- Single blank line between major sections
- Single blank line between each type definition
- Single blank line between each function/method/constructor
- Double blank lines are compacted to single
- No blank lines within const/var groups (only between groups)
