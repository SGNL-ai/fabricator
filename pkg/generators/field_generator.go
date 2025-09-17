package generators

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SGNL-ai/fabricator/pkg/util"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// FieldType represents the type of field being generated
type FieldType int

const (
	// FieldTypeUnknown is the default type for unclassified fields
	FieldTypeUnknown FieldType = iota
	// FieldTypeName represents a name field (person, entity, product, etc.)
	FieldTypeName
	// FieldTypeDescription represents a description or text field
	FieldTypeDescription
	// FieldTypeBoolean represents a true/false field
	FieldTypeBoolean
	// FieldTypeDate represents a date or timestamp field
	FieldTypeDate
	// FieldTypeStatus represents a status field
	FieldTypeStatus
	// FieldTypeGeneric represents a generic value field
	FieldTypeGeneric
)

// FieldRequest contains all information needed to generate a field value
type FieldRequest struct {
	// Entity and field information
	EntityID    string
	Header      string
	HeaderIndex int
	RowIndex    int

	// Special field flags
	IsUnique bool
}

// DetectFieldType determines the type of a field based on its header name
func DetectFieldType(header string) FieldType {
	headerLower := strings.ToLower(header)

	// Name field detection
	if headerLower == "name" || strings.HasSuffix(headerLower, "name") {
		return FieldTypeName
	}

	// Description field detection
	if headerLower == "description" || strings.HasSuffix(headerLower, "description") ||
		strings.Contains(headerLower, "desc") || strings.Contains(headerLower, "comment") ||
		strings.Contains(headerLower, "summary") || strings.Contains(headerLower, "notes") {
		return FieldTypeDescription
	}

	// Boolean field detection
	if headerLower == "valid" || headerLower == "enabled" || headerLower == "active" ||
		headerLower == "archived" || strings.HasSuffix(headerLower, "enabled") ||
		strings.HasSuffix(headerLower, "active") || strings.HasSuffix(headerLower, "valid") ||
		strings.HasSuffix(headerLower, "archived") {
		return FieldTypeBoolean
	}

	// Date field detection
	if strings.Contains(headerLower, "date") || strings.Contains(headerLower, "time") ||
		strings.Contains(headerLower, "created") || strings.Contains(headerLower, "updated") {
		return FieldTypeDate
	}

	// Status field detection
	if headerLower == "status" || strings.HasSuffix(headerLower, "status") {
		return FieldTypeStatus
	}

	// Default type
	return FieldTypeGeneric
}

// GenerateFieldValue creates a value for a field based on its type and context
func (g *CSVGenerator) GenerateFieldValue(req FieldRequest, fieldType FieldType) string {
	var value string

	switch fieldType {
	case FieldTypeName:
		value = g.generateNameField(req)
	case FieldTypeDescription:
		value = g.generateDescriptionField(req)
	case FieldTypeBoolean:
		value = g.generateBooleanField(req)
	case FieldTypeDate:
		value = g.generateDateField(req)
	case FieldTypeStatus:
		value = g.generateStatusField(req)
	case FieldTypeGeneric:
		value = g.generateGenericField(req)
	default:
		value = g.generateGenericField(req)
	}

	// For unique fields, ensure the value is unique
	if req.IsUnique && value != "" {
		value = g.ensureUniqueValue(req.EntityID, req.Header, value)
	}

	return value
}

// generateNameField creates values for name fields
func (g *CSVGenerator) generateNameField(req FieldRequest) string {
	// Use the entity ID directly from the request
	if req.EntityID != "" && g.EntityData[req.EntityID] != nil {
		entityName := strings.ToLower(g.EntityData[req.EntityID].EntityName)

		// Generate context-appropriate names based on entity type
		switch {
		// User-related entities should get person names
		case strings.Contains(entityName, "user") ||
			strings.Contains(entityName, "person") ||
			strings.Contains(entityName, "employee") ||
			strings.Contains(entityName, "customer"):
			return sanitizeName(gofakeit.Name())

		// Role-related entities should get job titles
		case strings.Contains(entityName, "role") ||
			strings.Contains(entityName, "job"):
			return sanitizeName(gofakeit.JobTitle())

		// Group-related entities should get department names
		case strings.Contains(entityName, "group") ||
			strings.Contains(entityName, "team") ||
			strings.Contains(entityName, "department"):
			departments := g.generatedValues["departments"]
			if len(departments) == 0 {
				return sanitizeName(gofakeit.Company() + " Department")
			}
			return sanitizeName(departments[req.RowIndex%len(departments)])

		// Product entities
		case strings.Contains(entityName, "product") ||
			strings.Contains(entityName, "item"):
			return sanitizeName(gofakeit.ProductName())

		// Company entities
		case strings.Contains(entityName, "company") ||
			strings.Contains(entityName, "organization"):
			return sanitizeName(gofakeit.Company())

		// Default to a company name
		default:
			return sanitizeName(gofakeit.Company())
		}
	}

	// Fallback if we couldn't determine the entity type
	return sanitizeName(gofakeit.Company())
}

// generateDescriptionField creates values for description fields
func (g *CSVGenerator) generateDescriptionField(req FieldRequest) string {
	return gofakeit.Sentence(util.CryptoRandInt(5) + 3) // 3-8 words
}

// generateBooleanField creates values for boolean fields
func (g *CSVGenerator) generateBooleanField(req FieldRequest) string {
	return strconv.FormatBool(req.RowIndex%2 == 0) // Alternate true/false
}

// generateDateField creates values for date fields
func (g *CSVGenerator) generateDateField(req FieldRequest) string {
	// Generate a date within the last 2 years
	minTime := time.Now().AddDate(-2, 0, 0)
	maxTime := time.Now()

	// Use gofakeit for better randomization
	date := gofakeit.DateRange(minTime, maxTime)

	// Format as YYYY-MM-DD
	return date.Format("2006-01-02")
}

// generateStatusField creates values for status fields
func (g *CSVGenerator) generateStatusField(req FieldRequest) string {
	statuses := g.generatedValues["status"]
	return statuses[req.RowIndex%len(statuses)]
}

// generateGenericField creates values for generic fields
func (g *CSVGenerator) generateGenericField(req FieldRequest) string {
	headerLower := strings.ToLower(req.Header)

	// More intelligent field type detection
	switch {
	// Numeric fields
	case strings.Contains(headerLower, "count") ||
		strings.Contains(headerLower, "number") ||
		strings.Contains(headerLower, "amount") ||
		strings.Contains(headerLower, "quantity"):
		if strings.Contains(headerLower, "price") || strings.Contains(headerLower, "cost") {
			// Generate a price with 2 decimal places (like $123.45)
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000))
		}
		return strconv.Itoa(gofakeit.Number(1, 1000))

	// Percentage fields
	case strings.Contains(headerLower, "percent") ||
		strings.Contains(headerLower, "rate"):
		return strconv.Itoa(gofakeit.Number(1, 100)) + "%"

	// Email fields
	case strings.Contains(headerLower, "email"):
		return gofakeit.Email()

	// Phone number fields
	case strings.Contains(headerLower, "phone"):
		return gofakeit.Phone()

	// URL fields
	case strings.Contains(headerLower, "url") ||
		strings.Contains(headerLower, "website") ||
		strings.Contains(headerLower, "link"):
		return gofakeit.URL()

	// Address fields
	case strings.Contains(headerLower, "address"):
		return gofakeit.Address().Address

	case strings.Contains(headerLower, "street"):
		return gofakeit.Street()

	case strings.Contains(headerLower, "city"):
		return gofakeit.City()

	case strings.Contains(headerLower, "state"):
		return gofakeit.State()

	case strings.Contains(headerLower, "zip") ||
		strings.Contains(headerLower, "postal"):
		return gofakeit.Zip()

	case strings.Contains(headerLower, "country"):
		return gofakeit.Country()

	// Color fields
	case strings.Contains(headerLower, "color") ||
		strings.Contains(headerLower, "colour"):
		return gofakeit.Color()

	// Time fields (excluding date fields)
	case strings.Contains(headerLower, "time") && !strings.Contains(headerLower, "date"):
		return gofakeit.Date().Format("15:04:05")

	// Password fields
	case strings.Contains(headerLower, "password"):
		return gofakeit.Password(true, true, true, true, false, 12)

	// Credit card fields
	case strings.Contains(headerLower, "credit") || strings.Contains(headerLower, "creditcard"):
		return gofakeit.CreditCardNumber(&gofakeit.CreditCardOptions{Types: []string{"visa", "mastercard"}})

	// IP address fields
	case headerLower == "ip" || strings.Contains(headerLower, "ipaddress"):
		return gofakeit.IPv4Address()

	// Comment/notes/summary fields (should be sentences)
	case strings.Contains(headerLower, "comment") ||
		strings.Contains(headerLower, "notes") ||
		strings.Contains(headerLower, "summary"):
		return gofakeit.Sentence(util.CryptoRandInt(5) + 3) // 3-8 words

	// UUID/GUID fields
	case strings.Contains(headerLower, "uuid") ||
		strings.Contains(headerLower, "guid"):
		return uuid.New().String()

	// Code fields
	case strings.Contains(headerLower, "code"):
		prefix := string([]rune(gofakeit.LetterN(3)))
		return strings.ToUpper(prefix) + "-" + strconv.Itoa(1000+req.RowIndex)

	// Default - generate a more interesting value using an adjective and noun
	default:
		return gofakeit.Word() + "_" + strconv.Itoa(req.RowIndex)
	}
}

// generateValue is used by some test cases and functions
func (g *CSVGenerator) generateValue(field string, index int) string {
	// Generate a more realistic value based on the field name
	fieldLower := strings.ToLower(field)

	// Try to infer the field type from its name
	switch {
	case strings.Contains(fieldLower, "color"):
		return gofakeit.Color()
	case strings.Contains(fieldLower, "currency"):
		return gofakeit.CurrencyShort()
	case strings.Contains(fieldLower, "job") || strings.Contains(fieldLower, "title"):
		return gofakeit.JobTitle()
	case strings.Contains(fieldLower, "company"):
		return gofakeit.Company()
	case strings.Contains(fieldLower, "product"):
		return gofakeit.ProductName()
	default:
		// Add a bit of randomness
		adjective := gofakeit.Adjective()
		noun := gofakeit.Noun()
		return adjective + "_" + noun + "_" + strconv.Itoa(index)
	}
}

// ensureUniqueValue ensures a value is unique within its entity and attribute scope
func (g *CSVGenerator) ensureUniqueValue(entityID, attrName, baseValue string) string {
	attrKey := entityID + ":" + attrName

	// If this is our first use of this attribute, initialize the map
	if g.usedUniqueValues[attrKey] == nil {
		g.usedUniqueValues[attrKey] = make(map[string]bool)
	}

	// For UUID values, just use a new UUID (they're already unique)
	if strings.Contains(strings.ToLower(attrName), "uuid") || strings.HasSuffix(strings.ToLower(attrName), "id") {
		uniqueVal := uuid.New().String()
		g.usedUniqueValues[attrKey][uniqueVal] = true
		return uniqueVal
	}

	// For other types, try to make the base value unique
	uniqueVal := baseValue
	attempt := 0

	// If this value is already used, append a suffix to make it unique
	for g.usedUniqueValues[attrKey][uniqueVal] && attempt < 1000 {
		// Add a suffix to make it unique
		suffix := "_" + strconv.Itoa(attempt)

		// If the value already has a numerical suffix, replace it
		if idx := strings.LastIndex(baseValue, "_"); idx != -1 {
			if _, err := strconv.Atoi(baseValue[idx+1:]); err == nil {
				uniqueVal = baseValue[:idx] + suffix
			} else {
				uniqueVal = baseValue + suffix
			}
		} else {
			uniqueVal = baseValue + suffix
		}

		attempt++
	}

	// Mark this value as used
	g.usedUniqueValues[attrKey][uniqueVal] = true
	return uniqueVal
}

// sanitizeName replaces commas and quotes in a name to avoid CSV parsing issues
func sanitizeName(name string) string {
	sanitized := strings.ReplaceAll(name, ",", "-")
	return strings.ReplaceAll(sanitized, "\"", "'")
}

// GenerateCommonValues pre-generates common test data values
func (g *CSVGenerator) generateCommonValues() {
	// Define common data types that can be used across any SOR

	// Common names for generic entities
	g.generatedValues["entityNames"] = []string{
		"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta",
		"Iota", "Kappa", "Lambda", "Mu", "Nu", "Xi", "Omicron", "Pi", "Rho",
		"Sigma", "Tau", "Upsilon", "Phi", "Chi", "Psi", "Omega",
	}

	// Common organization departments
	g.generatedValues["departments"] = []string{
		"Engineering", "Sales", "Marketing", "Finance", "HR", "Operations",
		"IT", "Legal", "Executive", "Support", "Research", "Development",
		"QA", "Product", "Design", "Customer Success", "Administration",
	}

	// Common person names
	g.generatedValues["firstNames"] = []string{
		"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
		"William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
		"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
		"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra", "Donald", "Ashley",
		"Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
		"Kenneth", "Dorothy", "Kevin", "Carol", "Brian", "Amanda", "George", "Melissa",
	}

	g.generatedValues["lastNames"] = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia",
		"Rodriguez", "Wilson", "Martinez", "Anderson", "Taylor", "Thomas", "Hernandez",
		"Moore", "Martin", "Jackson", "Thompson", "White", "Lopez", "Lee", "Gonzalez",
		"Harris", "Clark", "Lewis", "Robinson", "Walker", "Perez", "Hall", "Young",
		"Allen", "Sanchez", "Wright", "King", "Scott", "Green", "Baker", "Adams",
		"Nelson", "Hill", "Ramirez", "Campbell", "Mitchell", "Roberts", "Carter",
	}

	// Common permission values
	g.generatedValues["permissions"] = []string{
		"read", "write", "create", "delete", "admin", "view", "execute", "publish",
		"approve", "reject", "manage", "configure", "assign", "revoke", "audit",
	}

	// Status values
	g.generatedValues["status"] = []string{
		"active", "inactive", "pending", "approved", "rejected", "completed",
		"in_progress", "cancelled", "suspended", "archived", "draft",
	}

	// Common types
	g.generatedValues["types"] = []string{
		"primary", "secondary", "tertiary", "standard", "custom", "system", "user",
		"internal", "external", "public", "private", "shared", "restricted",
	}

	// Common keys
	g.generatedValues["keys"] = []string{
		"id", "name", "description", "type", "status", "category", "group",
		"code", "value", "priority", "order", "level", "tier", "version",
	}

	// Common expressions
	g.generatedValues["expressions"] = []string{
		"user.department == 'Engineering'",
		"item.status in ['active', 'pending']",
		"resource.owner == currentUser",
		"request.priority > 3",
		"document.classification != 'restricted'",
		"project.deadline < today()",
		"task.assignee == null",
	}
}
