// Package validate is for entity validation. It provides a Report that contains
// warnings and errors during validation as well as helper methods in the form of
// Assertion.
package validate

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/lefinal/nulls"
	"github.com/xeipuuv/gojsonschema"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Assertion returns a non-empty error message if the given value does not
// satisfy the requirements.
type Assertion[T any] func(val T) string

// AssertNotEmpty is an Assertion for the value not being equal to its empty value.
func AssertNotEmpty[T comparable]() Assertion[T] {
	return func(val T) string {
		var empty T
		if val == empty {
			return "required"
		}
		return ""
	}
}

// AssertNotNil is an Assertion that makes sure that the given value is not nil.
//
// Warning: Only use this on types that can actually be nil. Otherwise, consider
// using AssertNotEmpty.
func AssertNotNil[T any]() Assertion[T] {
	return func(val T) string {
		if reflect.ValueOf(val).IsNil() {
			return "required"
		}
		return ""
	}
}

// AssertDuration is an Assertion for the value being a valid duration string.
func AssertDuration[T string]() Assertion[T] {
	return func(val T) string {
		_, err := time.ParseDuration(string(val))
		if err != nil {
			return err.Error()
		}
		return ""
	}
}

// AssertIfOptionalStringSet checks the given Assertion-list if the value is set.
func AssertIfOptionalStringSet(assertions ...Assertion[string]) Assertion[nulls.String] {
	return func(val nulls.String) string {
		if len(assertions) == 0 {
			return "internal error: no assertions"
		}
		if !val.Valid {
			return ""
		}
		for _, assertion := range assertions {
			errMessage := assertion(val.String)
			if errMessage != "" {
				return errMessage
			}
		}
		return ""
	}
}

// AssertIfOptionalIntSet checks the given Assertion-list if the value is set.
func AssertIfOptionalIntSet(assertions ...Assertion[int]) Assertion[nulls.Int] {
	return func(val nulls.Int) string {
		if len(assertions) == 0 {
			return "internal error: no assertions"
		}
		if !val.Valid {
			return ""
		}
		for _, assertion := range assertions {
			errMessage := assertion(val.Int)
			if errMessage != "" {
				return errMessage
			}
		}
		return ""
	}
}

// AssertGreater is an Assertion that checks whether the given value is greater
// than the provided limit.
func AssertGreater[T cmp.Ordered](lower T) Assertion[T] {
	return func(val T) string {
		if val > lower {
			return ""
		}
		return fmt.Sprintf("should be greater than %v", lower)
	}
}

// AssertGreaterEq is an Assertion that checks whether the given value is greater
// or equal to the provided limit.
func AssertGreaterEq[T cmp.Ordered](lower T) Assertion[T] {
	return func(val T) string {
		if val >= lower {
			return ""
		}
		return fmt.Sprintf("should be greater or equal %v", lower)
	}
}

// AssertLess is an Assertion that checks whether the given value is less
// than the provided limit.
func AssertLess[T cmp.Ordered](lower T) Assertion[T] {
	return func(val T) string {
		if val < lower {
			return ""
		}
		return fmt.Sprintf("should be less than %v", lower)
	}
}

// AssertLessEq is an Assertion that checks whether the given value is less
// or equal to the provided limit.
func AssertLessEq[T cmp.Ordered](lower T) Assertion[T] {
	return func(val T) string {
		if val <= lower {
			return ""
		}
		return fmt.Sprintf("should be less or equal %v", lower)
	}
}

// AssertMaxStringLength asserts that the given string does not exceed the given
// maximum length.
func AssertMaxStringLength(maxLength int) Assertion[string] {
	return func(val string) string {
		if len(val) > maxLength {
			return fmt.Sprintf("length %d exceeds maximum length of %d", len(val), maxLength)
		}
		return ""
	}
}

// AssertASCIICharactersOnly is an Assertion that makes sure that the given
// string only consists of ASCII characters.
func AssertASCIICharactersOnly[T string]() Assertion[T] {
	return func(val T) string {
		for _, c := range val {
			if c > unicode.MaxASCII {
				return fmt.Sprintf("unallowed non-ascii character: %v", c)
			}
		}
		return ""
	}
}

// AssertAlphanumericCharactersOnlyWithExceptions is an Assertion that makes sure
// that the given string only consists of alphanumeric characters with the given
// allowed exceptions.
func AssertAlphanumericCharactersOnlyWithExceptions[T string](exceptChars []rune) Assertion[T] {
	return func(val T) string {
		// Start creating a regular expression pattern to match alphanumeric characters.
		pattern := `^[a-zA-Z0-9`
		// Add the exceptions to the pattern
		for _, e := range exceptChars {
			pattern += string(e)
		}
		// Finish up the regular expression and compile it.
		pattern += `]*$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Sprintf("compile regex for checking: %s", err.Error())
		}

		// Check if the string matches the pattern.
		if !re.MatchString(string(val)) {
			exceptCharsStr := make([]string, 0)
			for _, exceptChar := range exceptChars {
				exceptCharsStr = append(exceptCharsStr, string(exceptChar))
			}
			return fmt.Sprintf("only alphanumeric character with exceptions (%s) allowed", strings.Join(exceptCharsStr, ", "))
		}
		return ""
	}
}

// AssertLowercaseAlphanumericCharactersOnlyWithExceptions is an Assertion that
// makes sure that the given string only consists of lowercase alphanumeric
// characters with the given allowed exceptions.
func AssertLowercaseAlphanumericCharactersOnlyWithExceptions[T string](exceptChars []rune) Assertion[T] {
	return func(val T) string {
		// Start creating a regular expression pattern to match alphanumeric characters.
		pattern := `^[a-z0-9`
		// Add the exceptions to the pattern
		for _, e := range exceptChars {
			pattern += string(e)
		}
		// Finish up the regular expression and compile it.
		pattern += `]*$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Sprintf("compile regex for checking: %s", err.Error())
		}

		// Check if the string matches the pattern.
		if !re.MatchString(string(val)) {
			exceptCharsStr := make([]string, 0)
			for _, exceptChar := range exceptChars {
				exceptCharsStr = append(exceptCharsStr, string(exceptChar))
			}
			return fmt.Sprintf("only alphanumeric character with exceptions (%s) allowed", strings.Join(exceptCharsStr, ", "))
		}
		return ""
	}
}

// AssertOptionalValidJSONSchema is an Assertion that makes sure that the given
// optional JSON schema is valid.
func AssertOptionalValidJSONSchema() Assertion[nulls.JSONRawMessage] {
	return func(val nulls.JSONRawMessage) string {
		if !val.Valid {
			return ""
		}
		return AssertValidJSONSchema()(val.RawMessage)
	}
}

// AssertValidJSONSchema is an Assertion that makes sure that the given JSON
// schema is valid.
func AssertValidJSONSchema() Assertion[json.RawMessage] {
	return func(val json.RawMessage) string {
		schemaLoader := gojsonschema.NewSchemaLoader()
		schema := gojsonschema.NewBytesLoader(val)
		_, err := schemaLoader.Compile(schema)
		if err != nil {
			return fmt.Sprintf("invalid json schema: %s", err.Error())
		}
		return ""
	}
}

// AssertLowercaseCharactersOnly is an Assertion that makes sure that the given
// string only consists of lowercase characters.
func AssertLowercaseCharactersOnly[T string]() Assertion[T] {
	return func(val T) string {
		if string(val) != strings.ToLower(string(val)) {
			return "must consist of lowercase characters only"
		}
		return ""
	}
}

// AssertNoPrefix is an Assertion that makes sure that the given string does not
// have the given prefix.
func AssertNoPrefix[T string](prefix string) Assertion[T] {
	return func(val T) string {
		if strings.HasPrefix(string(val), prefix) {
			return fmt.Sprintf("must not have prefix %q", prefix)
		}
		return ""
	}
}

// AssertNoSuffix is an Assertion that makes sure that the given string does not
// have the given suffix.
func AssertNoSuffix[T string](suffix string) Assertion[T] {
	return func(val T) string {
		if strings.HasSuffix(string(val), suffix) {
			return fmt.Sprintf("must not have suffix %q", suffix)
		}
		return ""
	}
}

// AssertNoConsecutiveCharacter is an Assertion that makes sure that the given
// string does not consist of consecutive characters matching the given one.
func AssertNoConsecutiveCharacter[T string](targetCharacter rune) Assertion[T] {
	return func(val T) string {
		lastWasTargetCharacter := false
		for _, characterToCheck := range val {
			if characterToCheck == targetCharacter {
				if lastWasTargetCharacter {
					return fmt.Sprintf("no consecutive character '%v' allowed", targetCharacter)
				}
				lastWasTargetCharacter = true
			} else {
				lastWasTargetCharacter = false
			}
		}
		return ""
	}
}

// ForField checks the given Assertion-list on the provided value and reports the
// first encountered error, if any, to the Reporter.
func ForField[T any](reporter *Reporter, path *Path, val T, assertion Assertion[T], moreAssertions ...Assertion[T]) {
	assertions := append([]Assertion[T]{assertion}, moreAssertions...)
	reporter.NextField(path, val)
	for _, assertion := range assertions {
		errMessage := assertion(val)
		if errMessage != "" {
			reporter.Error(errMessage)
			return
		}
	}
}

// ForReporter checks the given Assertion-list on the provided value and reports the
// first encountered error, if any, to the Reporter.
func ForReporter[T any](reporter *Reporter, val T, assertion Assertion[T], moreAssertions ...Assertion[T]) {
	assertions := append([]Assertion[T]{assertion}, moreAssertions...)
	reporter.NextField(reporter.fieldPath, val)
	for _, assertion := range assertions {
		errMessage := assertion(val)
		if errMessage != "" {
			reporter.Error(errMessage)
			return
		}
	}
}

// JSONSchema validates the given value against the JSON schema. Errors for
// individual fields will be reported as individual issues in the returned
// Report.
func JSONSchema(path *Path, val json.RawMessage, jsonSchema json.RawMessage) *Report {
	reporter := NewReporter()
	if !json.Valid(val) {
		reporter.NextField(path, val)
		reporter.Error("invalid json")
		return reporter.Report()
	}
	schemaLoader := gojsonschema.NewBytesLoader(jsonSchema)
	configLoader := gojsonschema.NewBytesLoader(val)
	result, err := gojsonschema.Validate(schemaLoader, configLoader)
	if err != nil {
		reporter.NextField(path, val)
		reporter.Error(err.Error())
	} else if !result.Valid() {
		for _, resultError := range result.Errors() {
			reporter.NextField(path.Child(resultError.Field()), resultError.Value())
			reporter.Error(resultError.Description())
		}
	}
	return reporter.Report()
}

// validDockerImageNameRegex matches a valid Docker image name.
//
// Taken from Opus.
var validDockerImageNameRegex = regexp.MustCompile(`^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])(?:(?:\.(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]))+)?(?::[0-9]+)?/)?[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?(?:(?:/[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?)+)?(?::[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127})?$`)

// IsValidDockerImageName returns true if the given image name is valid.
func IsValidDockerImageName(imageName string) bool {
	return validDockerImageNameRegex.MatchString(imageName)
}
