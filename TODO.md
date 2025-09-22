# TODO

## Performance Optimization

**CRITICAL: O(n²) Performance Bug Found**

### Issue
- **Location**: `pkg/generators/model/entity.go:300-304` in `validateRow` function
- **Problem**: Uses O(n) linear search to check for duplicate primary keys on every row insertion
- **Impact**: Creates O(n²) total performance that makes large datasets unusable
- **Evidence**: CPU profiling shows `validateRow` consumes 68% of CPU time

### Performance Impact
- **1K records**: ~500K operations (acceptable)
- **20K records**: ~200M operations (31 seconds)
- **100K records**: ~5B operations (timeout)
- **1M records**: ~500B operations (completely unusable)

### Fix Required
Replace the O(n) loop with a hash map for O(1) duplicate detection:

```go
// Current O(n) approach:
for _, existingRow := range e.rows {
    if existingRow.values[pkName] == pkValue {
        return fmt.Errorf("duplicate value...")
    }
}

// Should be O(1) approach:
// Use a map[string]bool to track used primary key values
```

## Completed Items

✅ **Add comprehensive foreign key relationship testing**
✅ **Remove malformed sample.yaml and replace with proper examples**
✅ **Add CLI profiling support** (`--cpuprofile`, `--memprofile`)
✅ **Verify foreign key relationships work correctly** with both dotted notation and attributeAlias
✅ **Achieve 90.1% test coverage** with all packages above 80%
✅ **Remove debug output** and unreachable defensive code