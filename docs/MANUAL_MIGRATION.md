Next steps before using Agent mode:

## Pre-migration checklist

### Step 1: Fix lipgloss import path mismatch (required)

The crush packages (`internal/app/app.go` and `internal/agent/agent.go`) use `github.com/charmbracelet/lipgloss/v2`, but the module declares `github.com/charmbracelet/lipgloss/v2`.

Fix this:

1. Replace `github.com/charmbracelet/lipgloss/v2` with `github.com/charmbracelet/lipgloss/v2` in:
   - `internal/app/app.go` (line 19)
   - `internal/agent/agent.go` (line 30)

```bash
# Quick fix using sed (run from project root):
sed -i 's|github.com/charmbracelet/lipgloss/v2|github.com/charmbracelet/lipglossp/v2|g' internal/app/app.go internal/agent/agent.go
```

2. Verify the fix:
```bash
go mod tidy
go build ./...
```

### Step 2: Verify crush packages are available (already done)

From your terminal output, the crush packages are downloaded via `go mod`. They are available as external dependencies, not copied locally. This is fine—no manual copying needed.

### Step 3: Check compilation status

After fixing the lipgloss import:

```bash
# Check if everything compiles
go build ./cmd/review_helpers.go 2>&1 | head -20

# Or try building the whole project
go build ./...
```

Expected result: `cmd/review_helpers.go` should compile (the `message` package import should resolve).

---

## Summary: what to do before Agent mode

1. Fix lipgloss import paths (2 files: `internal/app/app.go`, `internal/agent/agent.go`)
2. Run `go mod tidy` to verify dependencies
3. Run `go build ./cmd/review_helpers.go` to verify compilation
4. Document any remaining compilation errors

After these steps, you can switch to Agent mode to continue the migration. The helper functions are ready; next is updating `cmd/review.go` to use the coordinator.

Should I provide the exact search-and-replace commands, or do you want to fix the imports and then continue?