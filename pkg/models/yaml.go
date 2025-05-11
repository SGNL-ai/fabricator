package models

// SORDefinition represents the top-level YAML structure defining a system of record
type SORDefinition struct {
	DisplayName               string                  `yaml:"displayName"`
	Description               string                  `yaml:"description"`
	Hostname                  string                  `yaml:"hostname"`
	DefaultSyncFrequency      string                  `yaml:"defaultSyncFrequency"`
	DefaultSyncMinInterval    int                     `yaml:"defaultSyncMinInterval"`
	DefaultApiCallFrequency   string                  `yaml:"defaultApiCallFrequency"`
	DefaultApiCallMinInterval int                     `yaml:"defaultApiCallMinInterval"`
	Type                      string                  `yaml:"type"`
	AdapterConfig             string                  `yaml:"adapterConfig"`
	Auth                      []map[string]AuthConfig `yaml:"auth"`
	Entities                  map[string]Entity       `yaml:"entities"`
	Relationships             map[string]Relationship `yaml:"relationships"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Username string `yaml:"username"`
}

// Entity represents a data entity in the SOR
type Entity struct {
	DisplayName        string      `yaml:"displayName"`
	ExternalId         string      `yaml:"externalId"`
	Description        string      `yaml:"description"`
	PagesOrderedById   bool        `yaml:"pagesOrderedById"`
	Attributes         []Attribute `yaml:"attributes"`
	EntityAlias        string      `yaml:"entityAlias"`
	SyncFrequency      string      `yaml:"syncFrequency,omitempty"`
	SyncMinInterval    int         `yaml:"syncMinInterval,omitempty"`
	ApiCallFrequency   string      `yaml:"apiCallFrequency,omitempty"`
	ApiCallMinInterval int         `yaml:"apiCallMinInterval,omitempty"`
}

// Attribute represents an attribute of an entity
type Attribute struct {
	Name           string `yaml:"name"`
	ExternalId     string `yaml:"externalId"`
	Description    string `yaml:"description"`
	Type           string `yaml:"type"`
	Indexed        bool   `yaml:"indexed"`
	UniqueId       bool   `yaml:"uniqueId"`
	AttributeAlias string `yaml:"attributeAlias"`
	List           bool   `yaml:"list"`
}

// RelationshipPath represents a path step in a relationship
type RelationshipPath struct {
	Relationship string `yaml:"relationship"`
	Direction    string `yaml:"direction"`
}

// Relationship represents relationships between entities
type Relationship struct {
	DisplayName   string             `yaml:"displayName"`
	Name          string             `yaml:"name"`
	FromAttribute string             `yaml:"fromAttribute,omitempty"`
	ToAttribute   string             `yaml:"toAttribute,omitempty"`
	Path          []RelationshipPath `yaml:"path,omitempty"`
}

// RelationshipLink represents a link between two entities for data generation purposes
type RelationshipLink struct {
	FromEntityID      string `json:"fromEntityID"`
	ToEntityID        string `json:"toEntityID"`
	FromAttribute     string `json:"fromAttribute"`
	ToAttribute       string `json:"toAttribute"`
	IsFromAttributeID bool   `json:"isFromAttributeID"` // Whether the from attribute is a unique ID
	IsToAttributeID   bool   `json:"isToAttributeID"`   // Whether the to attribute is a unique ID
}

// CSVData represents a structure to hold data for CSV file generation
type CSVData struct {
	ExternalId  string
	Headers     []string
	Rows        [][]string
	EntityName  string
	Description string
}
