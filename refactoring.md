# Fabricator Refactoring Plan

## Overview

This document outlines the refactoring plan for the Fabricator project, focusing on improving the architecture with a proper domain model and implementing test-driven development (TDD) techniques to ensure code quality and correctness.

## Domain Model Design

### 1. Graph Model

The Graph model will serve as the main coordinator that maintains entities and relationships, and provides methods to manage them.

#### Constructor
- `NewGraph(yamlModel *models.SORDefinition) (*Graph, error)` 
  - Creates a new Graph from the YAML model
  - Initializes internal entity map
  - Creates Entity objects for each YAML entity
  - Creates Relationship objects for connecting entities
  - Validates model integrity

#### Public Methods
- `GetEntity(id string) (*Entity, bool)` 
  - Gets an entity by ID with existence check
  
- `GetAllEntities() map[string]*Entity` 
  - Returns all entities in the graph
  
- `GetRelationship(id string) (*Relationship, bool)` 
  - Gets a relationship by ID with existence check
  
- `GetAllRelationships() []*Relationship` 
  - Returns all relationships in the graph
  
- `GetRelationshipsForEntity(entityID string) []*Relationship` 
  - Returns all relationships that involve a specific entity
  
- `GetTopologicalOrder() ([]string, error)` 
  - Returns entities in dependency order for generation
  - Implements stable sort for deterministic output

#### Internal Methods
- `createEntitiesFromYAML(yamlEntities map[string]models.Entity) error` 
  - Creates Entity objects from YAML model definition
  
- `createRelationshipsFromYAML(yamlRelationships map[string]models.Relationship) error` 
  - Creates Relationship objects from YAML model definition
  
- `validateGraph() error` 
  - Validates entire graph structure and integrity

### 2. Attribute Model

The Attribute model will represent entity attributes and their properties.

#### Constructor
- `newAttribute(name, externalID string, dataType string, isUnique bool, description string, parentEntity *Entity) *Attribute`
  - Creates a new attribute with all required properties
  - Not exported as only Entity should create attributes
  - Maintains reference to parent entity

#### Public Methods
- `GetName() string`
  - Returns attribute name
  
- `GetExternalID() string`
  - Returns attribute's external ID
  
- `GetDataType() string`
  - Returns attribute's data type
  
- `IsUnique() bool`
  - Returns whether attribute requires unique values
  
- `IsID() bool`
  - Returns whether attribute is an identifier
  
- `IsRelationship() bool`
  - Returns whether attribute is part of a relationship
  
- `GetParentEntity() *Entity`
  - Returns the parent entity this attribute belongs to
  
- `GetRelatedEntityID() string`
  - Returns the related entity ID if part of a relationship
  
- `GetRelatedAttribute() string`
  - Returns the related attribute name if part of a relationship

### 3. Entity Model

The Entity model will represent a data entity and manage its attributes and row data.

#### Constructor
- `newEntity(id, externalID, name string, description string, attributes []*Attribute) (*Entity, error)` 
  - Creates a new entity with basic properties and attributes
  - Performs validation of entity properties
  - Validates attribute uniqueness constraints
  - Sets parent entity reference on all attributes
  - Not exported as only Graph should create entities

#### Public Methods
- `GetID() string` 
  - Returns entity's internal ID
  
- `GetExternalID() string` 
  - Returns entity's external ID (used for CSV filenames)
  
- `GetName() string` 
  - Returns entity's display name
  
- `GetDescription() string` 
  - Returns entity's description
  
- `GetAttributes() []*Attribute` 
  - Returns all attributes in order
  
- `GetAttribute(name string) (*Attribute, bool)` 
  - Gets an attribute by name with existence check
  
- `GetPrimaryKey() *Attribute` 
  - Returns the single unique attribute that serves as primary key
  
- `GetNonUniqueAttributes() []*Attribute` 
  - Returns attributes not marked as unique
  
- `GetRelationshipAttributes() []*Attribute` 
  - Returns attributes involved in relationships
  
- `GetNonRelationshipAttributes() []*Attribute` 
  - Returns attributes not involved in relationships
  
- `GetRowCount() int` 
  - Returns the number of rows
  
- `AddRow(values map[string]string) error` 
  - Adds a new row with provided values
  - Validates uniqueness constraint for primary key
  - Validates foreign key values exist in related entities
  - Returns error if validation fails
  
- `ToCSV() *models.CSVData` 
  - Returns CSV representation of the entity

#### Internal Methods
- `validateRow(values map[string]string) error` 
  - Validates a row against entity constraints
  - Checks primary key uniqueness
  - Checks foreign key references validity
  
- `isUniqueValueUsed(value string) bool` 
  - Checks if a value is already used for the primary key
  
- `validateForeignKeyValue(attributeName string, value string) error`
  - Verifies that a foreign key value exists in the related entity
  - Returns error if the value doesn't exist in the related entity's primary key

### 4. Relationship Model

The Relationship model will represent a relationship between two entities and their attributes.

#### Constructor
- `newRelationship(id, name string, sourceEntity *Entity, targetEntity *Entity, sourceAttributeName string, targetAttributeName string) (*Relationship, error)` 
  - Creates a new relationship between entities
  - Validates that attributes exist on both entities
  - Validates that at least one side of the relationship has a unique attribute
  - Automatically determines cardinality based on attributes
  - Not exported as only Graph should create relationships
  - Returns error if validation fails

#### Public Methods
- `GetID() string` 
  - Returns relationship's ID
  
- `GetName() string` 
  - Returns relationship's name
  
- `GetSourceEntity() *Entity` 
  - Returns source entity
  
- `GetTargetEntity() *Entity` 
  - Returns target entity
  
- `GetSourceAttribute() *Attribute` 
  - Returns source attribute
  
- `GetTargetAttribute() *Attribute` 
  - Returns target attribute
  
- `GetCardinality() string` 
  - Returns relationship cardinality (1:1, 1:N, N:1)
  
- `IsOneToOne() bool` 
  - Returns true if relationship is 1:1
  
- `IsOneToMany() bool` 
  - Returns true if relationship is 1:N
  
- `IsManyToOne() bool` 
  - Returns true if relationship is N:1

#### Internal Methods
- `validateAttributes() error` 
  - Validates source and target attributes exist
  - Validates that at least one side has a unique attribute
  
- `determineCardinality()` 
  - Analyzes attributes to determine cardinality
  - Sets relationship cardinality based on uniqueness

## Refactoring Steps

1. **Create Domain Model**
   - Implement attribute model
   - Implement entity model with validation
   - Implement relationship model with cardinality detection
   - Implement graph model with YAML parsing and topological ordering

2. **Integrate with CSV Generator**
   - Create adapter between model and CSV generator
   - Update relationship handling to use new models
   - Update validation to use new models

3. **Refactor CSV Generator**
   - Make it a coordinator with steps in sub-packages
   - Design pipeline architecture for generation steps
   - Extract core generation steps into dedicated modules

4. **Redesign Data Generation Flow**
   - Phase 1: Generate identifier fields in topological order
   - Phase 2: Establish relationship structure between entities
   - Phase 3: Fill in remaining non-relationship fields

5. **Improve APIs and Interfaces**
   - Create clear interfaces between generation phases
   - Add explicit support for dependencies in field generation
   - Rename modules for clarity and purpose

## Test-Driven Development Approach

We are implementing a comprehensive Test-Driven Development (TDD) approach to guide our refactoring effort. This methodology ensures we build robust, well-tested components from the ground up.

### TDD Workflow

Our TDD process follows this iterative pattern:

1. **Write Failing Tests First**:
   - Begin by writing tests that clearly define expected behavior
   - Ensure tests fail initially (Red phase)
   - Document requirements through test cases
   - Include edge cases and error handling scenarios

2. **Implement Minimal Code**:
   - Write just enough code to make tests pass (Green phase)
   - Focus on correctness rather than optimization
   - Implement one requirement at a time
   - Validate functionality through test results

3. **Refactor with Confidence**:
   - Improve code structure and readability (Refactor phase)
   - Maintain test coverage during refactoring
   - Extract common patterns and remove duplication
   - Optimize performance where needed
   
4. **Repeat the Cycle**:
   - Continue with the next requirement or feature
   - Add new tests before adding new functionality
   - Maintain the Red-Green-Refactor discipline

### Testing Strategy

Our testing approach emphasizes several key principles:

1. **Clean, Modular Test Design**:
   - Keep test logic and test data separate
   - Group related tests into focused test functions
   - Use descriptive test and test case names
   - Apply the DRY principle within reason for test code
   - Write tests that serve as documentation

2. **Test Structure**:
   - Use table-driven tests to handle multiple scenarios
   - Properly initialize and clean up test resources
   - Use consistent assertion patterns
   - Favor testify assertions for clear error messages
   - Keep tests fast and deterministic

3. **Test Fixtures**:
   - Create reusable test fixtures for common scenarios
   - Use helper functions to construct test data
   - Define small, focused test datasets
   - Initialize test objects with valid default values
   - Customize only relevant properties for each test

4. **Model Validation Testing**:
   - Verify constructor parameter validation
   - Test boundary conditions and edge cases
   - Ensure error cases return appropriate errors
   - Validate public method contracts
   - Test state consistency after operations

5. **Relationship Testing**:
   - Validate relationship cardinality detection
   - Test circular dependency handling
   - Verify foreign key consistency rules
   - Test relationship traversal functionality
   - Ensure proper parent/child relationships

### Implementation Progress

We've successfully applied this TDD approach to our initial work:

1. **Attribute Model**:
   - Created comprehensive tests for all Attribute functionality
   - Implemented relationship handling and validation
   - Ensured proper parent/child reference integrity
   - Validated getter methods and relationship state

2. **Entity Model** (in progress):
   - Developed test suite covering core Entity functionality
   - Focused on attribute management and validation
   - Included tests for key constraints (unique attributes)
   - Created tests for row management and validation
   - Added CSV output validation

This TDD methodology is helping us build a robust domain model that correctly implements all business rules while maintaining high test coverage. It also serves as living documentation of system behavior, making it easier for future developers to understand the codebase.

The initial test implementation phase has already highlighted several design improvements we're incorporating into our models:

1. **Attribute-Entity Relationship**:
   - Attributes maintain a reference to their parent entity
   - Ensures consistent bidirectional navigation
   - Enables relationship validation from either direction

2. **Entity Construction**:
   - Entities receive their attributes at construction time
   - Prevents inconsistent state by validating at creation
   - Enforces "exactly one unique attribute" rule from the start
   - Simplifies validation by front-loading constraints

3. **Row Management**:
   - Added explicit validation of row data against entity schema
   - Implemented primary key uniqueness enforcement
   - Created foreign key validation between related entities
   - Built data integrity checking into the add operation

This test-driven approach is helping us build a more robust, maintainable, and correct implementation than the previous version.

## Object Model Integration Architecture

The domain model components will work together according to the following architecture:

1. **Graph Initialization**:
   - We instantiate a `Graph` object with the YAML structure (`models.SORDefinition`)
   - The `Graph` serves as the top-level container and coordinator for all entities and relationships
   - It's responsible for building and maintaining the complete model from the YAML definition

2. **Entity and Attribute Construction**:
   - The `Graph` creates all `Entity` instances and their `Attribute` instances
   - Each `Entity` has a reference back to its parent `Graph`, enabling lookups to other entities
   - All entities are initialized with their attributes before relationships are established

3. **Dependency Management**:
   - The `Graph` incorporates the topological sorting logic from `generators/dependency_graph`
   - This determines the correct processing order based on entity relationships
   - Entities that are referenced by others will be processed first during data generation

4. **Relationship Establishment**:
   - For each relationship in the YAML, the `Graph`:
     - Locates the source entity
     - Calls `addRelationship(...)` on that entity
     - The entity uses its `Graph` reference to find the target entity
     - It then creates a `Relationship` object connecting the entities
   - This creates a complete relationship structure with bidirectional navigation

The resulting object graph provides:
- A fully connected network of entities and relationships
- Bidirectional navigation capabilities (entity → relationship → entity)
- Type-safe access to related objects
- Built-in validation of relationship constraints
- Support for foreign key validation during data generation
- A foundation for maintaining referential integrity across generated data

This architecture enables a more maintainable and correct implementation by explicitly modeling the relationships between entities, rather than relying on implicit connections through attribute names or values.

## Constructor Pattern

We follow a consistent constructor pattern for all domain model objects that prioritizes validation, separation of concerns, and maintainability:

### Four-Step Constructor Pattern

1. **Object Creation**
   - Create a minimal object instance with the provided parameters
   - Initialize the object's basic state, but don't perform complex operations yet
   - Store input parameters for later validation and setup

2. **Validation**
   - Validate all input parameters before proceeding with complex operations
   - Check basic invariants (non-nil pointers, non-empty required strings, etc.)
   - Return clear, specific error messages for validation failures
   - Use a dedicated `validate()` method to encapsulate validation logic

3. **Setup**
   - Set up internal state and relationships, only after validation passes
   - Establish references to other objects in the domain model
   - Initialize internal data structures and collections
   - Use dedicated setup methods for complex initialization logic

4. **Business Logic**
   - Apply any business logic required to finalize the object's state
   - Calculate derived properties (e.g., relationship cardinality)
   - Set default values where appropriate
   - Ensure the object is in a valid, consistent state before returning

### Benefits of this Pattern

- **Clear Separation of Concerns**: Each phase has a distinct responsibility
- **Early Validation**: Catches errors before performing expensive operations
- **Testability**: Each phase can be tested independently
- **Maintainability**: Easier to understand, modify, and extend
- **Predictable Construction**: Consistent sequence across all domain objects
- **Better Error Messages**: Errors can point to specific issues in specific phases
- **Fail-Fast Principle**: Invalid objects cannot be constructed

This pattern is applied consistently across all domain model objects including Entity, Attribute, Relationship, and Graph, providing a unified approach to object construction and validation.