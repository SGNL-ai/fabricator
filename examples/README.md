# Fabricator Example YAML Files

This directory contains example YAML files for testing and demonstrating Fabricator functionality.

## Valid Examples

### okta.sgnl.yaml
Complete Okta System of Record definition with users, groups, applications, and group membership relationships.
- **Use case**: Real-world example from Okta integration
- **Features**: Multiple entities, direct relationships, path-based relationships

### okta.sgnl.exported.yaml
Exported version of Okta SOR with additional metadata.
- **Use case**: Example of exported SOR configuration
- **Features**: Same structure as okta.sgnl.yaml with export metadata

### salesforce.sgnl.yaml
Complete Salesforce System of Record definition with accounts, users, cases, and various relationships.
- **Use case**: Real-world example from Salesforce integration
- **Features**: Complex entity relationships, multiple entity types

## Invalid Examples (for Testing)

### bad-relationships.yaml
Tests various relationship validation errors.
- **Errors demonstrated**:
  - Relationship referencing non-existent attribute
  - Relationship referencing non-existent entity
  - Path-based relationship referencing non-existent relationship
- **Expected behavior**: Parser validation should fail with detailed error messages

### invalid-displayname-relationships.yaml
Tests relationship attribute reference format validation.
- **Error demonstrated**: Incorrect attribute reference format in relationships
- **Example**:
  - Entity has `displayName: User` and `externalId: ons-profile-read/user`
  - Correct format: `fromAttribute: ons-profile-read/user.profileId` (uses ExternalId)
- **Expected behavior**: Parser validation should fail and show available attribute patterns

### invalid-enum-types.yaml
Tests invalid enumeration type values.
- **Error demonstrated**: Invalid enum values in attribute definitions
- **Expected behavior**: Schema validation should fail

### invalid-malformed-yaml.yaml
Tests malformed YAML syntax.
- **Error demonstrated**: YAML syntax errors
- **Expected behavior**: YAML parser should fail with syntax error

### invalid-missing-fields.yaml
Tests missing required fields.
- **Error demonstrated**: Missing required fields in entity or relationship definitions
- **Expected behavior**: Schema validation should fail

## Relationship Attribute Reference Formats

Fabricator supports two formats for referencing attributes in relationships:

### 1. Attribute Alias (Recommended)
When an attribute has an `attributeAlias`, use the alias directly:
```yaml
attributes:
  - name: id
    externalId: id
    attributeAlias: user-id-unique-alias
relationships:
  my_rel:
    fromAttribute: user-id-unique-alias
```

### 2. ExternalId Format
Use the entity's `externalId` combined with the attribute's `externalId`:
```yaml
entities:
  User:
    externalId: ons-profile-read/user
    attributes:
      - name: id
        externalId: id
relationships:
  my_rel:
    fromAttribute: ons-profile-read/user.id
```

**Resolution Priority**: Fabricator first tries to resolve `attributeAlias`, then falls back to `externalId` format.

## Testing Examples

```bash
# Test valid example
fabricator -f examples/okta.sgnl.yaml -n 100

# Test invalid relationship references (should show helpful error)
fabricator -f examples/invalid-displayname-relationships.yaml -n 10

# Test bad relationships (should show validation errors)
fabricator -f examples/bad-relationships.yaml -n 10
```
