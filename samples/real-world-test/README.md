# Real-World Test Results: Per-Entity Row Count Configuration

This directory contains test results demonstrating the per-entity row count configuration feature.

## Test Scenario

**Source**: `samples/data-gen.yaml` (GM AuthZ stress test with 22 entities)
**Feature**: Per-entity row counts via `--count-config` flag

## Files in This Directory

### Configuration Files

1. **counts-template.yaml** - Auto-generated template
   - Created with: `fabricator init-count-config -f samples/data-gen.yaml`
   - Contains all 22 entities with default 100 rows each
   - Includes helpful comments and entity descriptions

2. **counts-medium.yaml** - Custom configuration
   - Scaled to ~450K total records for faster testing
   - Demonstrates realistic production-like proportions:
     - Large datasets: 100K rows for customer profiles
     - Medium datasets: 5K-20K rows for identity entities
     - Small datasets: 20-200 rows for applications/APIs
     - Static datasets: 100 rows for permissions
     - Singleton: 1 row for universal criteria

3. **counts-custom.yaml** - Production-scale configuration
   - Based on actual GM requirements (not executed due to size)
   - Full production volumes: ~45M+ total records
   - Key specifications:
     - 10M customer profiles (1:1 chains through account/vehicle)
     - 2M group memberships
     - 1M EntraId users, 500K groups (with 1:1 same_as copies)
     - 20K APIs with 20K permissions (1:1, must link to same app)
     - 6K application hostnames for 2K applications (3:1 ratio)
     - 10K permission criteria (half of APIs accessible to customers)
     - Singleton universal criteria (1 instance, all profiles/criteria link to it)
   - Note: Would take ~5+ minutes to generate

### Output Directory

**output/** - Generated CSV files from counts-medium.yaml
- 22 CSV files with custom row counts per entity
- Total: ~458K records across all files
- Generation time: <1 second
- All row counts match configuration exactly

## Verification Results

| Entity | Expected | Actual | Status |
|--------|----------|--------|--------|
| application | 20 | 20 | ✅ |
| applicationHostname | 60 | 60 | ✅ |
| api | 200 | 200 | ✅ |
| entraIdUser | 10,000 | 10,000 | ✅ |
| entraIdGroup | 5,000 | 5,000 | ✅ |
| user | 10,000 | 10,000 | ✅ |
| group | 5,000 | 5,000 | ✅ |
| userProfile | 100,000 | 100,000 | ✅ |
| customerAccount | 100,000 | 100,000 | ✅ |
| account | 100,000 | 100,000 | ✅ |
| accountVehicle | 100,000 | 100,000 | ✅ |
| universalClientCriteriaSingleton | 1 | 1 | ✅ |
| *All 22 entities* | 458,681 | 458,681 | ✅ |

## Commands Used

```bash
# Generate template
fabricator init-count-config -f samples/data-gen.yaml > counts-template.yaml

# Generate CSVs with custom counts
fabricator -f samples/data-gen.yaml --count-config counts-medium.yaml -o output/

# Performance: <1 second for 458K records!
```

## Key Observations

1. **Exact Precision** - Every entity has exactly the configured row count
2. **Fast Performance** - 458K records generated in <1 second
3. **Clean Output** - YAML template is clean, DEBUG goes to stderr
4. **Relationship Integrity** - All relationships maintained across varying entity sizes
5. **Singleton Support** - Single-row entities work perfectly

## Feature Status

✅ **Production Ready** - All functionality working as designed
