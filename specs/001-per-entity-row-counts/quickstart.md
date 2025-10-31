# Quickstart: Per-Entity Row Count Configuration

**Feature**: 001-per-entity-row-counts
**Audience**: Fabricator users who want different row counts per entity
**Time**: 5 minutes

## What This Feature Does

Before this feature, fabricator generated the same number of rows for every entity:
```bash
fabricator -f sor.yaml -n 100  # All entities get 100 rows
```

Now you can specify different row counts per entity:
```bash
fabricator -f sor.yaml --count-config counts.yaml  # Custom counts per entity
```

**Use Cases**:
- Realistic test data distributions (1000 users, 10 departments)
- Performance testing with varied entity sizes
- Edge case testing with minimal data for specific entities

## Quick Start

### Step 1: Generate a Configuration Template

```bash
fabricator init-count-config -f your-sor.yaml > counts.yaml
```

This creates a template with all entities from your SOR file:

```yaml
# Row count configuration for fabricator
# Generated from: your-sor.yaml
# Last updated: 2025-10-30

# Entity: users
# Description: User accounts
users: 100

# Entity: groups
# Description: User groups
groups: 100

# Entity: permissions
# Description: Access permissions
permissions: 100
```

### Step 2: Edit Row Counts

Open `counts.yaml` and customize the numbers:

```yaml
# Row count configuration for fabricator
# Generated from: your-sor.yaml
# Last updated: 2025-10-30

# Entity: users
# Description: User accounts
users: 1000  # ← Changed to 1000

# Entity: groups
# Description: User groups
groups: 50   # ← Changed to 50

# Entity: permissions
# Description: Access permissions
permissions: 200  # ← Changed to 200
```

### Step 3: Generate CSV Files

```bash
fabricator -f your-sor.yaml --count-config counts.yaml -o output/
```

Check the output:
```bash
wc -l output/*.csv
```

Output:
```
    1001 output/users.csv       # 1000 rows + 1 header
      51 output/groups.csv      # 50 rows + 1 header
     201 output/permissions.csv # 200 rows + 1 header
```

**Done!** You now have CSV files with custom row counts per entity.

## Common Scenarios

### Scenario 1: Large Users, Small Reference Data

```yaml
users: 10000
organizations: 100
roles: 10
permissions: 50
```

Generate:
```bash
fabricator -f sor.yaml --count-config counts.yaml -o output/
```

### Scenario 2: Minimal Test Data

```yaml
users: 5
groups: 2
permissions: 3
```

Perfect for quick integration tests with small datasets.

### Scenario 3: Performance Testing

```yaml
transactions: 1000000
customers: 100000
products: 10000
orders: 500000
```

Test system behavior with large datasets while keeping reference data manageable.

### Scenario 4: Mixed (Some Custom, Some Default)

```yaml
# Only specify entities that need custom counts
users: 5000
orders: 25000
# Other entities will default to 100 rows
```

Run with:
```bash
fabricator -f sor.yaml --count-config counts.yaml -o output/
```

Entities not in `counts.yaml` get 100 rows by default.

## Backward Compatibility

**Old way still works!** Existing commands are unchanged:

```bash
# This still works exactly as before
fabricator -f sor.yaml -n 100 -o output/
```

All entities get 100 rows (uniform count).

## Advanced Usage

### Combine with Relationship Validation

```bash
fabricator -f sor.yaml --count-config counts.yaml --validate -o output/
```

Fabricator will warn if row counts create impossible relationships:

```
Warning: Cardinality violation - Relationship 'user_groups' (one-to-many)
100 groups require 1+ user each but only 50 users exist.
Some groups will have no associated users.
```

CSV files are still generated with best-effort relationship assignment.

### Redirect Template Output

```bash
# Generate template directly to file
fabricator init-count-config -f sor.yaml > my-counts.yaml

# Generate and immediately edit
fabricator init-count-config -f sor.yaml | vim -
```

### Check Configuration Without Generating Data

```bash
# Dry-run: validate configuration file
fabricator -f sor.yaml --count-config counts.yaml --validate-only
```

This checks:
- Configuration file syntax
- Entity names match SOR
- Row counts are valid integers
- Relationship cardinality feasibility

## Troubleshooting

### Error: "Entity 'foo' in count configuration not found in SOR YAML"

**Cause**: Your `counts.yaml` references an entity that doesn't exist in the SOR file.

**Fix**: Check entity names in your SOR file:
```bash
grep "external_id:" your-sor.yaml
```

Remove or rename the problematic entity in `counts.yaml`.

### Error: "Cannot use both -n flag and --count-config file"

**Cause**: You provided both `-n 100` and `--count-config counts.yaml`.

**Fix**: Choose one approach:
- Use `-n` for uniform counts across all entities, OR
- Use `--count-config` for per-entity counts

### Error: "Invalid count for entity 'users': 0"

**Cause**: Row count must be a positive integer (>0).

**Fix**: Change `users: 0` to `users: 1` or higher in your config file.

### Warning: Cardinality Violations

**Cause**: Your row counts make it impossible to satisfy relationship cardinality.

**Example**: 100 departments each need 1+ employee, but only 50 employees specified.

**Impact**: CSV files are still generated. Relationships are assigned best-effort. Some departments will have no employees.

**Fix**: Adjust row counts to satisfy cardinality:
```yaml
departments: 100
employees: 150  # Increased to satisfy 1+ per department
```

## Tips & Best Practices

### 1. Start with Template

Always generate a template first rather than writing config from scratch:
```bash
fabricator init-count-config -f sor.yaml > counts.yaml
```

This ensures correct entity names and provides helpful comments.

### 2. Version Control Your Configs

Save your `counts.yaml` files in git:
```bash
git add counts-dev.yaml counts-staging.yaml counts-prod.yaml
git commit -m "Add row count configurations for different environments"
```

### 3. Use Descriptive Filenames

```bash
counts-small.yaml    # For quick tests (10-50 rows)
counts-medium.yaml   # For integration tests (100-1000 rows)
counts-large.yaml    # For performance tests (10K-1M rows)
```

### 4. Document Rationale

Add comments to your config files explaining unusual counts:

```yaml
users: 10000
# Large user base to test pagination performance

groups: 5
# Minimal groups - testing single-group edge case

permissions: 1000
# Many permissions to test authorization matrix complexity
```

### 5. Check Relationship Feasibility

Before generating large datasets, validate with smaller counts first:

```yaml
# Test with small counts
users: 10
groups: 5
```

Run and check for warnings. Then scale up:

```yaml
# Production-like counts
users: 10000
groups: 500
```

## Next Steps

- **Specification**: Read [spec.md](./spec.md) for complete functional requirements
- **Data Model**: See [data-model.md](./data-model.md) for implementation details
- **Schema**: Reference [contracts/count-config-schema.yaml](./contracts/count-config-schema.yaml) for validation rules
- **Implementation**: View [tasks.md](./tasks.md) (after `/speckit.tasks` command) for development tasks

## Examples

### Example 1: E-commerce System

```yaml
# E-commerce test data
customers: 5000
orders: 15000
products: 1000
order_items: 45000
categories: 50
reviews: 8000
```

### Example 2: Identity Management

```yaml
# Identity/access management
users: 10000
groups: 200
roles: 50
permissions: 500
group_memberships: 25000
role_assignments: 15000
```

### Example 3: Content Management

```yaml
# CMS test data
authors: 100
articles: 5000
categories: 30
tags: 200
comments: 15000
media_assets: 3000
```

## FAQ

**Q: Can I omit entities from the config file?**
A: Yes. Entities not in the config will use the default count (100 rows).

**Q: What's the maximum row count?**
A: No hard limit, but very large counts (>1M) may cause out-of-memory errors. This is your responsibility to manage.

**Q: Can I use expressions like "10% of users"?**
A: No. Only literal integer values are supported. See spec.md "Out of Scope" for rationale.

**Q: Does this work with the existing `-a` (auto-cardinality) flag?**
A: Yes. All existing flags are compatible with `--count-config`.

**Q: Can I use JSON instead of YAML?**
A: No. Only YAML format is supported for consistency with SOR files.

**Q: Will this break my existing scripts?**
A: No. All existing commands work unchanged. This feature is purely additive.

## Getting Help

- **Issues**: Report bugs at [github.com/SGNL-ai/fabricator/issues](https://github.com/SGNL-ai/fabricator/issues)
- **Documentation**: See main [README.md](../../../README.md)
- **Examples**: Check [examples/](../../../examples/) directory for sample SOR files

---

**Version**: 1.0.0 | **Last Updated**: 2025-10-30 | **Feature**: 001-per-entity-row-counts
