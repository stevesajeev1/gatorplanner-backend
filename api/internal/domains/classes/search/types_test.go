package search

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		check     func(t *testing.T, rule Rule)
	}{
		{
			name:  "valid general text field",
			input: `{"field":"course_name","type":"text","filter":"equal","value":"Data Structures"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseName, rule.Field)
				assert.Equal(t, FieldTypeText, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)
				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Data Structures", *rule.TextValue)
			},
		},
		{
			name:  "valid checked text field",
			input: `{"field":"meet_type","type":"text","filter":"equal","value":"Primarily Classroom"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldClassMeetType, rule.Field)
				assert.Equal(t, FieldTypeText, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)
				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Primarily Classroom", *rule.TextValue)
			},
		},
		{
			name:  "valid checked text field gen ed",
			input: `{"field":"course_gen_eds","type":"text","filter":"equal","value":"Composition"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseGenEds, rule.Field)
				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Composition", *rule.TextValue)
			},
		},
		{
			name:  "valid checked text field quest",
			input: `{"field":"course_quest","type":"text","filter":"equal","value":"Quest 1"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseQuest, rule.Field)
				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Quest 1", *rule.TextValue)
			},
		},
		{
			name:  "valid checked text field meet days",
			input: `{"field":"meet_times.days","type":"text","filter":"equal","value":"M"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseMeetDays, rule.Field)
				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "M", *rule.TextValue)
			},
		},

		{
			name:      "missing value",
			input:     `{"field":"course_name","type":"text","filter":"equal"}`,
			wantError: true,
		},
		{
			name:      "invalid text field",
			input:     `{"field":"invalid","type":"text","filter":"equal","value":"test"}`,
			wantError: true,
		},
		{
			name:      "invalid text filter",
			input:     `{"field":"course_name","type":"text","filter":"greater","value":"test"}`,
			wantError: true,
		},
		{
			name:      "invalid checked meet type",
			input:     `{"field":"meet_type","type":"text","filter":"equal","value":"invalid"}`,
			wantError: true,
		},
		{
			name:      "invalid checked gen ed",
			input:     `{"field":"course_gen_eds","type":"text","filter":"equal","value":"invalid"}`,
			wantError: true,
		},
		{
			name:      "invalid checked quest",
			input:     `{"field":"course_quest","type":"text","filter":"equal","value":"invalid"}`,
			wantError: true,
		},
		{
			name:      "invalid checked meet day",
			input:     `{"field":"meet_times.days","type":"text","filter":"equal","value":"X"}`,
			wantError: true,
		},
		{
			name:      "text value has wrong type",
			input:     `{"field":"course_name","type":"text","filter":"equal","value":123}`,
			wantError: true,
		},

		{
			name:  "valid number equal",
			input: `{"field":"course_level","type":"number","filter":"equal","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseLevel, rule.Field)
				assert.Equal(t, FieldTypeNumber, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)
				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3000.0, *rule.NumberValue)
			},
		},
		{
			name:  "valid number greater",
			input: `{"field":"course_level","type":"number","filter":"greater","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3000.0, *rule.NumberValue)
			},
		},
		{
			name:  "valid number less",
			input: `{"field":"course_level","type":"number","filter":"less","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3000.0, *rule.NumberValue)
			},
		},
		{
			name:  "valid number greater or equal",
			input: `{"field":"course_level","type":"number","filter":"greaterOrEqual","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3000.0, *rule.NumberValue)
			},
		},
		{
			name:  "valid number less or equal",
			input: `{"field":"course_level","type":"number","filter":"lessOrEqual","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3000.0, *rule.NumberValue)
			},
		},
		{
			name:      "invalid number field",
			input:     `{"field":"course_name","type":"number","filter":"equal","value":3000}`,
			wantError: true,
		},
		{
			name:      "invalid number filter",
			input:     `{"field":"course_level","type":"number","filter":"includes","value":3000}`,
			wantError: true,
		},
		{
			name:      "number value has wrong type",
			input:     `{"field":"course_level","type":"number","filter":"equal","value":"3000"}`,
			wantError: true,
		},

		{
			name:  "valid boolean true",
			input: `{"field":"course_is_lab","type":"boolean","filter":"equal","value":true}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseIsLab, rule.Field)
				assert.Equal(t, FieldTypeBoolean, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)
				require.NotNil(t, rule.BooleanValue)
				assert.True(t, *rule.BooleanValue)
			},
		},
		{
			name:  "valid boolean false",
			input: `{"field":"course_is_lab","type":"boolean","filter":"notEqual","value":false}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.BooleanValue)
				assert.False(t, *rule.BooleanValue)
			},
		},
		{
			name:      "invalid boolean field",
			input:     `{"field":"course_name","type":"boolean","filter":"equal","value":true}`,
			wantError: true,
		},
		{
			name:      "invalid boolean filter",
			input:     `{"field":"course_is_lab","type":"boolean","filter":"greater","value":true}`,
			wantError: true,
		},
		{
			name:      "boolean value has wrong type",
			input:     `{"field":"course_is_lab","type":"boolean","filter":"equal","value":"true"}`,
			wantError: true,
		},

		{
			name:      "invalid field type",
			input:     `{"field":"course_name","type":"invalid","filter":"equal","value":"test"}`,
			wantError: true,
		},
		{
			name:      "malformed json",
			input:     `{"field":"course_name","type":"text","filter":"equal","value":`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule Rule

			err := json.Unmarshal([]byte(tt.input), &rule)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.check != nil {
				tt.check(t, rule)
			}
		})
	}
}

func TestFilterUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		check     func(t *testing.T, filter Filter)
	}{
		{
			name: "single rule",
			input: `{
				"glue": "and",
				"rules": [
					{
						"field": "course_is_lab",
						"type": "boolean",
						"filter": "equal",
						"value": true
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 1)

				rule, ok := filter.Rules[0].(Rule)
				require.True(t, ok)

				require.NotNil(t, rule.BooleanValue)
				assert.True(t, *rule.BooleanValue)
			},
		},
		{
			name: "or rules",
			input: `{
				"glue": "or",
				"rules": [
					{
						"field": "course_department",
						"type": "text",
						"filter": "equal",
						"value": "CIS"
					},
					{
						"field": "course_department",
						"type": "text",
						"filter": "equal",
						"value": "COP"
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueOr, filter.Glue)
				require.Len(t, filter.Rules, 2)

				for i, expected := range []string{"CIS", "COP"} {
					rule, ok := filter.Rules[i].(Rule)
					require.True(t, ok)
					require.NotNil(t, rule.TextValue)
					assert.Equal(t, expected, *rule.TextValue)
				}
			},
		},
		{
			name: "nested filter",
			input: `{
				"glue": "and",
				"rules": [
					{
						"field": "course_is_lab",
						"type": "boolean",
						"filter": "equal",
						"value": true
					},
					{
						"glue": "or",
						"rules": [
							{
								"field": "course_department",
								"type": "text",
								"filter": "equal",
								"value": "CIS"
							},
							{
								"field": "course_department",
								"type": "text",
								"filter": "equal",
								"value": "COP"
							}
						]
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 2)

				_, ok := filter.Rules[0].(Rule)
				assert.True(t, ok)

				nested, ok := filter.Rules[1].(Filter)
				require.True(t, ok)

				assert.Equal(t, GlueOr, nested.Glue)
				require.Len(t, nested.Rules, 2)
			},
		},
		{
			name:      "invalid glue",
			input:     `{"glue":"invalid","rules":[]}`,
			wantError: true,
		},
		{
			name:      "empty rules",
			input:     `{"glue":"and","rules":[]}`,
			wantError: true,
		},
		{
			name: "invalid node",
			input: `{
				"glue": "and",
				"rules": [
					{
						"foo": "bar"
					}
				]
			}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter Filter

			err := json.Unmarshal([]byte(tt.input), &filter)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.check != nil {
				tt.check(t, filter)
			}
		})
	}
}
