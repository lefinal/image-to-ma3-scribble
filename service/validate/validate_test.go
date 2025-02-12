package validate

import (
	"encoding/json"
	"github.com/lefinal/nulls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testValidator[T any](value T, assertion Assertion[T], moreAssertions ...Assertion[T]) func(reporter *Reporter) {
	return func(reporter *Reporter) {
		ForField(reporter, NewPath(""), value, assertion, moreAssertions...)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expectErr bool
		validate  func(reporter *Reporter)
	}{
		// Assert not empty.
		{
			name:      "assert not empty int is set",
			expectErr: false,
			validate:  testValidator[int](42, AssertNotEmpty[int]()),
		},
		{
			name:      "assert not empty int not set",
			expectErr: true,
			validate:  testValidator[int](0, AssertNotEmpty[int]()),
		},
		{
			name:      "assert not empty uint8 is set",
			expectErr: false,
			validate:  testValidator[uint8](42, AssertNotEmpty[uint8]()),
		},
		{
			name:      "assert not empty uint8 not set",
			expectErr: true,
			validate:  testValidator[uint8](0, AssertNotEmpty[uint8]()),
		},
		{
			name:      "assert not empty string is set",
			expectErr: false,
			validate:  testValidator[string]("Hello World!", AssertNotEmpty[string]()),
		},
		{
			name:      "assert not empty string not set",
			expectErr: true,
			validate:  testValidator[string]("", AssertNotEmpty[string]()),
		},

		// Assert not nil.
		{
			name:      "assert not nil ok",
			expectErr: false,
			validate: testValidator([]byte(`Hello World!`),
				AssertNotNil[[]byte]()),
		},
		{
			name:      "assert not nil ok 2",
			expectErr: false,
			validate: testValidator(NewReporter(),
				AssertNotNil[*Reporter]()),
		},
		{
			name:      "assert not nil ok 3",
			expectErr: false,
			validate: testValidator([]string{},
				AssertNotNil[[]string]()),
		},
		{
			name:      "assert not nil ok 4",
			expectErr: false,
			validate: testValidator(json.RawMessage{},
				AssertNotNil[json.RawMessage]()),
		},
		{
			name:      "assert not nil invalid",
			expectErr: true,
			validate: testValidator(nil,
				AssertNotNil[json.RawMessage]()),
		},

		// Assert duration.
		{
			name:      "assert duration ok",
			expectErr: false,
			validate: testValidator[string]("10s",
				AssertDuration()),
		},
		{
			name:      "assert duration empty",
			expectErr: true,
			validate: testValidator[string]("",
				AssertDuration()),
		},
		{
			name:      "assert duration invalid",
			expectErr: true,
			validate: testValidator[string]("Hello World!",
				AssertDuration()),
		},

		// Assert if optional string set.
		{
			name:      "assert if optional string set no assertions",
			expectErr: true,
			validate: testValidator[nulls.String](nulls.String{},
				AssertIfOptionalStringSet()),
		},
		{
			name:      "assert if optional string set not set",
			expectErr: false,
			validate: testValidator[nulls.String](nulls.String{},
				AssertIfOptionalStringSet(
					AssertDuration(),
				)),
		},
		{
			name:      "assert if optional string set is set invalid",
			expectErr: true,
			validate: testValidator[nulls.String](nulls.NewString("Hello World!"),
				AssertIfOptionalStringSet(
					AssertDuration(),
				)),
		},
		{
			name:      "assert if optional string set is set ok",
			expectErr: false,
			validate: testValidator[nulls.String](nulls.NewString("100ms"),
				AssertIfOptionalStringSet(
					AssertDuration(),
				)),
		},

		// Assert if optional int set.
		{
			name:      "assert if optional int set no assertions",
			expectErr: true,
			validate: testValidator[nulls.Int](nulls.Int{},
				AssertIfOptionalIntSet()),
		},
		{
			name:      "assert if optional int set not set",
			expectErr: false,
			validate: testValidator[nulls.Int](nulls.Int{},
				AssertIfOptionalIntSet(
					AssertGreater(0),
					AssertGreater(42),
				)),
		},
		{
			name:      "assert if optional int set is set invalid",
			expectErr: true,
			validate: testValidator[nulls.Int](nulls.NewInt(5),
				AssertIfOptionalIntSet(
					AssertGreater(0),
					AssertGreater(42),
				)),
		},
		{
			name:      "assert if optional int set is set ok",
			expectErr: false,
			validate: testValidator[nulls.Int](nulls.NewInt(100),
				AssertIfOptionalIntSet(
					AssertGreater(0),
					AssertGreater(42),
				)),
		},

		// Assert greater.
		{
			name:      "assert greater but equal",
			expectErr: true,
			validate: testValidator(123,
				AssertGreater(123),
			),
		},
		{
			name:      "assert greater but less",
			expectErr: true,
			validate: testValidator(30,
				AssertGreater(123),
			),
		},
		{
			name:      "assert greater ok",
			expectErr: false,
			validate: testValidator(140,
				AssertGreater(123),
			),
		},

		// Assert greater equals.
		{
			name:      "assert greater equals and is equal",
			expectErr: false,
			validate: testValidator(123,
				AssertGreaterEq(123),
			),
		},
		{
			name:      "assert greater equals but less",
			expectErr: true,
			validate: testValidator(30,
				AssertGreaterEq(123),
			),
		},
		{
			name:      "assert greater equals ok",
			expectErr: false,
			validate: testValidator(140,
				AssertGreaterEq(123),
			),
		},

		// Assert less.
		{
			name:      "assert less but equal",
			expectErr: true,
			validate: testValidator(123,
				AssertLess(123),
			),
		},
		{
			name:      "assert less but greater",
			expectErr: true,
			validate: testValidator(500,
				AssertLess(123),
			),
		},
		{
			name:      "assert less ok",
			expectErr: false,
			validate: testValidator(15,
				AssertLess(123),
			),
		},

		// Assert less equals.
		{
			name:      "assert less equals and is equal",
			expectErr: false,
			validate: testValidator(123,
				AssertLessEq(123),
			),
		},
		{
			name:      "assert less equals but greater",
			expectErr: true,
			validate: testValidator(500,
				AssertLessEq(123),
			),
		},
		{
			name:      "assert less equals ok",
			expectErr: false,
			validate: testValidator(43,
				AssertLessEq(123),
			),
		},

		// Assert max string length.
		{
			name:      "assert max string length ok",
			expectErr: false,
			validate: testValidator("abc",
				AssertMaxStringLength(5),
			),
		},
		{
			name:      "assert max string length eq",
			expectErr: false,
			validate: testValidator("abc",
				AssertMaxStringLength(3),
			),
		},
		{
			name:      "assert max string length exceeded",
			expectErr: true,
			validate: testValidator("abc",
				AssertMaxStringLength(2),
			),
		},

		// Assert ASCII characters only.
		{
			name:      "assert ascii characters only ok",
			expectErr: false,
			validate: testValidator("abc",
				AssertASCIICharactersOnly()),
		},
		{
			name:      "assert ascii characters only empty",
			expectErr: false,
			validate: testValidator("",
				AssertASCIICharactersOnly()),
		},
		{
			name:      "assert ascii characters only invalid",
			expectErr: true,
			validate: testValidator("abcö",
				AssertASCIICharactersOnly()),
		},

		// Assert alphanumeric characters only with exceptions.
		{
			name:      "assert alphanumeric characters only with exceptions ok",
			expectErr: false,
			validate: testValidator("abc",
				AssertAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert alphanumeric characters only with exceptions uppercase",
			expectErr: false,
			validate: testValidator("ACB13",
				AssertAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert alphanumeric characters only with exceptions ok with exception char",
			expectErr: false,
			validate: testValidator("ab_c2",
				AssertAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert alphanumeric characters only with exceptions ok with exception char at start",
			expectErr: false,
			validate: testValidator("_abc1",
				AssertAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert alphanumeric characters only with exceptions ok with exception char at end",
			expectErr: false,
			validate: testValidator("ab2c_",
				AssertAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert alphanumeric characters only with exceptions empty",
			expectErr: false,
			validate: testValidator("",
				AssertAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert alphanumeric characters only with exceptions invalid",
			expectErr: true,
			validate: testValidator("a!b1c",
				AssertAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert alphanumeric characters only with exceptions invalid 2",
			expectErr: true,
			validate: testValidator("a_b2c!",
				AssertAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},

		// Assert lowercase alphanumeric characters only with exceptions.
		{
			name:      "assert lowercase alphanumeric characters only with exceptions ok",
			expectErr: false,
			validate: testValidator("abc",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions uppercase",
			expectErr: true,
			validate: testValidator("ACB13",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions ok with exception char",
			expectErr: false,
			validate: testValidator("ab_c2",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions ok with exception char at start",
			expectErr: false,
			validate: testValidator("_abc1",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions ok with exception char at end",
			expectErr: false,
			validate: testValidator("ab2c_",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions empty",
			expectErr: false,
			validate: testValidator("",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions invalid",
			expectErr: true,
			validate: testValidator("a!b1c",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions(nil)),
		},
		{
			name:      "assert lowercase alphanumeric characters only with exceptions invalid 2",
			expectErr: true,
			validate: testValidator("a_b2c!",
				AssertLowercaseAlphanumericCharactersOnlyWithExceptions([]rune{'_'})),
		},

		// Assert optional valid JSON schema.
		{
			name:      "assert optional valid json schema ok",
			expectErr: false,
			validate: testValidator(nulls.NewJSONRawMessage(json.RawMessage(`{"type": "string"}`)),
				AssertOptionalValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema empty",
			expectErr: false,
			validate: testValidator(nulls.JSONRawMessage{Valid: false},
				AssertOptionalValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema nil",
			expectErr: true,
			validate: testValidator(nulls.NewJSONRawMessage(nil),
				AssertOptionalValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema invalid json",
			expectErr: true,
			validate: testValidator(nulls.NewJSONRawMessage(json.RawMessage(`{invalid`)),
				AssertOptionalValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema invalid schema",
			expectErr: true,
			validate: testValidator(nulls.NewJSONRawMessage(json.RawMessage(`{"type": "xstring"}`)),
				AssertOptionalValidJSONSchema()),
		},

		// Assert valid JSON schema.
		{
			name:      "assert optional valid json schema ok",
			expectErr: false,
			validate: testValidator(json.RawMessage(`{"type": "string"}`),
				AssertValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema nil",
			expectErr: true,
			validate: testValidator(nil,
				AssertValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema invalid json",
			expectErr: true,
			validate: testValidator(json.RawMessage(`{invalid`),
				AssertValidJSONSchema()),
		},
		{
			name:      "assert optional valid json schema invalid schema",
			expectErr: true,
			validate: testValidator(json.RawMessage(`{"type": "xstring"}`),
				AssertValidJSONSchema()),
		},

		// Assert lowercase characters only.
		{
			name:      "assert lowercase characters only ok",
			expectErr: false,
			validate: testValidator("hello world",
				AssertLowercaseCharactersOnly()),
		},
		{
			name:      "assert lowercase characters only ok with special characters",
			expectErr: false,
			validate: testValidator("hello world 123 !?_+",
				AssertLowercaseCharactersOnly()),
		},
		{
			name:      "assert lowercase characters only empty",
			expectErr: false,
			validate: testValidator("",
				AssertLowercaseCharactersOnly()),
		},
		{
			name:      "assert lowercase characters only invalid",
			expectErr: true,
			validate: testValidator("Hello World",
				AssertLowercaseCharactersOnly()),
		},

		// Assert no prefix.
		{
			name:      "assert no prefix ok single",
			expectErr: false,
			validate: testValidator("hello world",
				AssertNoPrefix("a")),
		},
		{
			name:      "assert no prefix ok multi",
			expectErr: false,
			validate: testValidator("hello world",
				AssertNoPrefix("ola")),
		},
		{
			name:      "assert no prefix invalid single",
			expectErr: true,
			validate: testValidator("a cookie",
				AssertNoPrefix("a")),
		},
		{
			name:      "assert no prefix invalid multi",
			expectErr: true,
			validate: testValidator("hello world",
				AssertNoPrefix("hello")),
		},

		// Assert no suffix.
		{
			name:      "assert no suffix ok single",
			expectErr: false,
			validate: testValidator("hello world",
				AssertNoSuffix("a")),
		},
		{
			name:      "assert no suffix ok multi",
			expectErr: false,
			validate: testValidator("hello world",
				AssertNoSuffix("ola")),
		},
		{
			name:      "assert no suffix invalid single",
			expectErr: true,
			validate: testValidator("a cookie",
				AssertNoSuffix("e")),
		},
		{
			name:      "assert no suffix invalid multi",
			expectErr: true,
			validate: testValidator("hello world",
				AssertNoSuffix("ld")),
		},

		// Assert no consecutive characters.
		{
			name:      "assert no consecutive characters ok with no target char",
			expectErr: false,
			validate: testValidator("abc",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with single target char",
			expectErr: false,
			validate: testValidator("ab_c",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with multiple target chars",
			expectErr: false,
			validate: testValidator("a_b_c",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with single target char at start",
			expectErr: false,
			validate: testValidator("_abc",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with single target char at end",
			expectErr: false,
			validate: testValidator("abc_",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with consecutive target chars",
			expectErr: true,
			validate: testValidator("ab__c",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with single and consecutive target chars",
			expectErr: true,
			validate: testValidator("a_b__c",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with consecutive target chars at start",
			expectErr: true,
			validate: testValidator("__abc",
				AssertNoConsecutiveCharacter('_')),
		},
		{
			name:      "assert no consecutive characters ok with consecutive target chars at end",
			expectErr: true,
			validate: testValidator("abc__",
				AssertNoConsecutiveCharacter('_')),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter()
			tt.validate(reporter)
			errList := reporter.Report().Errors
			if tt.expectErr {
				assert.NotEmpty(t, errList)
			} else {
				assert.Empty(t, errList)
			}
		})
	}
}

func TestJSONSchema(t *testing.T) {
	tests := []struct {
		name         string
		val          json.RawMessage
		jsonSchema   json.RawMessage
		expectErrors []Issue
	}{
		{
			name:       "empty",
			val:        json.RawMessage(`{}`),
			jsonSchema: json.RawMessage(`{}`),
		},
		{
			name: "ok",
			val: json.RawMessage(`{
  "hello": "world"
}`),
			jsonSchema: json.RawMessage(`{
  "properties": {
    "hello": {
      "type": "string"
    }
  }
}`),
		},
		{
			name: "invalid type",
			val: json.RawMessage(`{
  "hello": 1
}`),
			jsonSchema: json.RawMessage(`{
  "properties": {
    "hello": {
      "type": "string"
    }
  }
}`),
			expectErrors: []Issue{
				{
					Field:    "my-config.hello",
					BadValue: json.Number("1"),
					Detail:   "Invalid type. Expected: string, given: integer",
				},
			},
		},
		{
			name: "multiple invalid",
			val: json.RawMessage(`{
  "a": 1,
  "b": 2,
  "c": 3
}`),
			jsonSchema: json.RawMessage(`{
  "properties": {
    "a": {
      "type": "number"
    },
    "b": {
      "type": "string"
    },
    "c": {
      "type": "number",
      "minimum": 10
    }
  },
  "required": [
    "a",
    "d"
  ]
}`),
			expectErrors: []Issue{
				{
					Field:    "my-config.b",
					BadValue: json.Number("2"),
					Detail:   "Invalid type. Expected: string, given: integer",
				},
				{
					Field:    "my-config.c",
					BadValue: json.Number("3"),
					Detail:   "Must be greater than or equal to 10",
				},
				{
					Field: "my-config.(root)",
					BadValue: map[string]any{
						"a": json.Number("1"),
						"b": json.Number("2"),
						"c": json.Number("3"),
					},
					Detail: "d is required",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := JSONSchema(NewPath("my-config"), tt.val, tt.jsonSchema)
			assert.ElementsMatch(t, tt.expectErrors, report.Errors)
		})
	}
}

func Test_IsValidDockerImageName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Image string
		Valid bool
	}{
		{
			Image: "myimage",
			Valid: true,
		},
		{
			Image: "my_image",
			Valid: true,
		},
		{
			Image: "my-image",
			Valid: true,
		},
		{
			Image: "my.image",
			Valid: true,
		},
		{
			Image: "myimage:tag",
			Valid: true,
		},
		{
			Image: "my/image",
			Valid: true,
		},
		{
			Image: "registry.example.com:5000/my/image",
			Valid: true,
		},
		{
			Image: "my/image:latest",
			Valid: true,
		},
		{
			Image: "MY_IMAGE",
			Valid: false,
		},
		{
			Image: "my/image!",
			Valid: false,
		},
		{
			Image: "my/image:v1.0.0",
			Valid: true,
		},
		{
			Image: "my/image:v1.0",
			Valid: true,
		},
		{
			Image: "my/image:1.0",
			Valid: true,
		},
		{
			Image: "my/image:latest",
			Valid: true,
		},
		{
			Image: "my/image:MY_TAG",
			Valid: true,
		},
		{
			Image: "my/image:tag-with-hyphen",
			Valid: true,
		},
		{
			Image: "my/image:tag_with_underscore",
			Valid: true,
		},
		{
			Image: "my/image:tag.with.period",
			Valid: true,
		},
		{
			Image: "my/image:",
			Valid: false,
		},
		{
			Image: "my/image:tag:with:colon",
			Valid: false,
		},
		{
			Image: "my/image:tag with space",
			Valid: false,
		},
		{
			Image: "my/image:tag@with@invalid@char",
			Valid: false,
		},
		{
			Image: "my/image:tag/with/slash",
			Valid: false,
		},
		{
			Image: "my/image:tag\\with\\backslash",
			Valid: false,
		},
		{
			Image: "my/image:tag*with*asterisk",
			Valid: false,
		},
		{
			Image: "my/image:tag?with?question?mark",
			Valid: false,
		},
		{
			Image: "my/image:tag#with#hash",
			Valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Image, func(t *testing.T) {
			assert.Equal(t, tt.Valid, IsValidDockerImageName(tt.Image))
		})
	}
}
