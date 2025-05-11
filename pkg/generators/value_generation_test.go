package generators

import (
	"strconv"
	"strings"
	"testing"
)

func TestGenerateGenericValue(t *testing.T) {
	generator := NewCSVGenerator("test_output", 5)
	generator.generateCommonValues() // Initialize common values

	// Test all the different field types that generateGenericValue can handle
	testCases := []struct {
		name       string
		fieldName  string
		index      int
		validateFn func(string) bool
		errorMsg   string
	}{
		{
			name:      "Price field",
			fieldName: "itemPrice",
			index:     1,
			validateFn: func(s string) bool {
				// Just check that it's not empty
				return len(s) > 0
			},
			errorMsg: "Price fields should not be empty",
		},
		{
			name:      "Count field",
			fieldName: "count",
			index:     2,
			validateFn: func(s string) bool {
				// Should be numeric
				_, err := strconv.Atoi(s)
				return err == nil
			},
			errorMsg: "Count fields should be numeric",
		},
		{
			name:      "Number field",
			fieldName: "number_of_items",
			index:     3,
			validateFn: func(s string) bool {
				// Should be numeric
				_, err := strconv.Atoi(s)
				return err == nil
			},
			errorMsg: "Number fields should be numeric",
		},
		{
			name:      "Amount field",
			fieldName: "totalAmount",
			index:     4,
			validateFn: func(s string) bool {
				// Should be numeric
				_, err := strconv.Atoi(s)
				return err == nil
			},
			errorMsg: "Amount fields should be numeric",
		},
		{
			name:      "Quantity field",
			fieldName: "quantity",
			index:     5,
			validateFn: func(s string) bool {
				// Should be numeric
				_, err := strconv.Atoi(s)
				return err == nil
			},
			errorMsg: "Quantity fields should be numeric",
		},
		{
			name:      "Percentage field",
			fieldName: "percentage",
			index:     6,
			validateFn: func(s string) bool {
				return strings.Contains(s, "%")
			},
			errorMsg: "Percentage fields should contain %",
		},
		{
			name:      "Rate field",
			fieldName: "rate",
			index:     7,
			validateFn: func(s string) bool {
				return strings.Contains(s, "%")
			},
			errorMsg: "Rate fields should contain %",
		},
		{
			name:      "Email field",
			fieldName: "email",
			index:     8,
			validateFn: func(s string) bool {
				return strings.Contains(s, "@")
			},
			errorMsg: "Email fields should contain @",
		},
		{
			name:      "Phone field",
			fieldName: "phone",
			index:     9,
			validateFn: func(s string) bool {
				return len(s) >= 10
			},
			errorMsg: "Phone fields should be at least 10 characters",
		},
		{
			name:      "URL field",
			fieldName: "url",
			index:     10,
			validateFn: func(s string) bool {
				return strings.Contains(s, "://")
			},
			errorMsg: "URL fields should contain ://",
		},
		{
			name:      "Website field",
			fieldName: "website",
			index:     11,
			validateFn: func(s string) bool {
				return strings.Contains(s, "://")
			},
			errorMsg: "Website fields should contain ://",
		},
		{
			name:      "Link field",
			fieldName: "link",
			index:     12,
			validateFn: func(s string) bool {
				return strings.Contains(s, "://")
			},
			errorMsg: "Link fields should contain ://",
		},
		{
			name:      "Username field",
			fieldName: "username",
			index:     13,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Username fields should not be empty",
		},
		{
			name:      "Password field",
			fieldName: "password",
			index:     14,
			validateFn: func(s string) bool {
				return len(s) >= 8
			},
			errorMsg: "Password fields should be at least 8 characters",
		},
		{
			name:      "Address field",
			fieldName: "address",
			index:     15,
			validateFn: func(s string) bool {
				// Address should contain spaces (multiple words)
				return strings.Contains(s, " ")
			},
			errorMsg: "Address fields should contain spaces (multiple words)",
		},
		{
			name:      "Street field",
			fieldName: "street",
			index:     16,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Street fields should not be empty",
		},
		{
			name:      "City field",
			fieldName: "city",
			index:     17,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "City fields should not be empty",
		},
		{
			name:      "State field",
			fieldName: "state",
			index:     18,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "State fields should not be empty",
		},
		{
			name:      "Zip field",
			fieldName: "zip",
			index:     19,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Zip fields should not be empty",
		},
		{
			name:      "Postal field",
			fieldName: "postalCode",
			index:     20,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Postal fields should not be empty",
		},
		{
			name:      "Country field",
			fieldName: "country",
			index:     21,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Country fields should not be empty",
		},
		{
			name:      "First Name field",
			fieldName: "firstName",
			index:     22,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "First name fields should not be empty",
		},
		{
			name:      "Last Name field",
			fieldName: "lastName",
			index:     23,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Last name fields should not be empty",
		},
		{
			name:      "Code field",
			fieldName: "code",
			index:     24,
			validateFn: func(s string) bool {
				return strings.Contains(s, "-")
			},
			errorMsg: "Code fields should contain a dash (-)",
		},
		{
			name:      "IP field",
			fieldName: "ip",
			index:     25,
			validateFn: func(s string) bool {
				return strings.Count(s, ".") == 3
			},
			errorMsg: "IP fields should contain 3 dots (IPv4 format)",
		},
		{
			name:      "Credit Card field",
			fieldName: "creditCard",
			index:     26,
			validateFn: func(s string) bool {
				return len(s) > 10
			},
			errorMsg: "Credit Card fields should be longer than 10 characters",
		},
		{
			name:      "Time field",
			fieldName: "time",
			index:     27,
			validateFn: func(s string) bool {
				return strings.Count(s, ":") == 2
			},
			errorMsg: "Time fields should have format HH:MM:SS",
		},
		{
			name:      "Color field",
			fieldName: "color",
			index:     28,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Color fields should not be empty",
		},
		{
			name:      "Department field",
			fieldName: "department",
			index:     29,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Department fields should not be empty",
		},
		{
			name:      "Product field",
			fieldName: "product",
			index:     30,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Product fields should not be empty",
		},
		{
			name:      "Comment field",
			fieldName: "comment",
			index:     31,
			validateFn: func(s string) bool {
				return strings.Contains(s, " ")
			},
			errorMsg: "Comment fields should contain spaces (be a sentence)",
		},
		{
			name:      "Summary field",
			fieldName: "summary",
			index:     32,
			validateFn: func(s string) bool {
				return strings.Contains(s, " ")
			},
			errorMsg: "Summary fields should contain spaces (be a sentence)",
		},
		{
			name:      "Notes field",
			fieldName: "notes",
			index:     33,
			validateFn: func(s string) bool {
				return strings.Contains(s, " ")
			},
			errorMsg: "Notes fields should contain spaces (be a sentence)",
		},
		{
			name:      "UUID field",
			fieldName: "uuid",
			index:     34,
			validateFn: func(s string) bool {
				return strings.Count(s, "-") == 4
			},
			errorMsg: "UUID fields should contain 4 dashes",
		},
		{
			name:      "GUID field",
			fieldName: "guid",
			index:     35,
			validateFn: func(s string) bool {
				return strings.Count(s, "-") == 4
			},
			errorMsg: "GUID fields should contain 4 dashes",
		},
		{
			name:      "Default field",
			fieldName: "someGenericField",
			index:     36,
			validateFn: func(s string) bool {
				return strings.Contains(s, "_")
			},
			errorMsg: "Default fields should contain an underscore",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := generator.generateGenericValue(tc.fieldName, tc.index)

			// Value should not be empty
			if value == "" {
				t.Errorf("Generated value for %s is empty", tc.fieldName)
			}

			// Value should match the validation function
			if !tc.validateFn(value) {
				t.Errorf("Generated value '%s' for field '%s' is invalid: %s",
					value, tc.fieldName, tc.errorMsg)
			}
		})
	}
}

func TestGenerateValue(t *testing.T) {
	generator := NewCSVGenerator("test_output", 5)
	generator.generateCommonValues() // Initialize common values

	// Test the different field types that generateValue handles
	testCases := []struct {
		name       string
		fieldName  string
		index      int
		validateFn func(string) bool
		errorMsg   string
	}{
		{
			name:      "Color field",
			fieldName: "color",
			index:     1,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Color fields should not be empty",
		},
		{
			name:      "Currency field",
			fieldName: "currency",
			index:     2,
			validateFn: func(s string) bool {
				return len(s) == 3 // Most currency codes are 3 characters
			},
			errorMsg: "Currency fields should be 3 characters",
		},
		{
			name:      "Job Title field",
			fieldName: "jobTitle",
			index:     3,
			validateFn: func(s string) bool {
				return len(s) > 0 // Just check it's not empty
			},
			errorMsg: "Job title fields should not be empty",
		},
		{
			name:      "Company field",
			fieldName: "company",
			index:     4,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Company fields should not be empty",
		},
		{
			name:      "Product field",
			fieldName: "product",
			index:     5,
			validateFn: func(s string) bool {
				return len(s) > 0
			},
			errorMsg: "Product fields should not be empty",
		},
		{
			name:      "Default field",
			fieldName: "someField",
			index:     6,
			validateFn: func(s string) bool {
				return strings.Count(s, "_") >= 1
			},
			errorMsg: "Default fields should contain underscores",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := generator.generateValue(tc.fieldName, tc.index)

			// Value should not be empty
			if value == "" {
				t.Errorf("Generated value for %s is empty", tc.fieldName)
			}

			// Value should match the validation function
			if !tc.validateFn(value) {
				t.Errorf("Generated value '%s' for field '%s' is invalid: %s",
					value, tc.fieldName, tc.errorMsg)
			}
		})
	}
}
