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
			name:  "valid general text",
			input: `{"field":"course_name","type":"text","filter":"equal","value":"Computer Science"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseName, rule.Field)
				assert.Equal(t, FieldTypeText, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Computer Science", *rule.TextValue)
			},
		},
		{
			name:  "valid checked text",
			input: `{"field":"meet_type","type":"text","filter":"equal","value":"Hybrid"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldClassMeetType, rule.Field)
				assert.Equal(t, FieldTypeText, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Hybrid", *rule.TextValue)
			},
		},
		{
			name:  "valid not equal text",
			input: `{"field":"course_name","type":"text","filter":"notEqual","value":"Computer Science"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldFilterNotEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Computer Science", *rule.TextValue)
			},
		},
		{
			name:      "invalid text field",
			input:     `{"field":"invalid","type":"text","filter":"equal","value":"test"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid text filter",
			input:     `{"field":"course_name","type":"text","filter":"greater","value":"test"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid meet type",
			input:     `{"field":"meet_type","type":"text","filter":"equal","value":"INVALID"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid gen ed",
			input:     `{"field":"course_gen_eds","type":"text","filter":"equal","value":"INVALID"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid quest",
			input:     `{"field":"course_quest","type":"text","filter":"equal","value":"INVALID"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid meet day",
			input:     `{"field":"meet_times.days","type":"text","filter":"equal","value":"INVALID"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:  "valid number",
			input: `{"field":"course_level","type":"number","filter":"greater","value":3000}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseLevel, rule.Field)
				assert.Equal(t, FieldTypeNumber, rule.Type)
				assert.Equal(t, FieldFilterGreater, rule.Filter)

				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, float64(3000), *rule.NumberValue)
			},
		},
		{
			name:  "valid decimal number",
			input: `{"field":"course_credits","type":"number","filter":"lessOrEqual","value":3.5}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseCredits, rule.Field)
				assert.Equal(t, FieldFilterLessOrEqual, rule.Filter)

				require.NotNil(t, rule.NumberValue)
				assert.Equal(t, 3.5, *rule.NumberValue)
			},
		},
		{
			name:  "valid period",
			input: `{"field":"meet_times.period_start","type":"number","filter":"greaterOrEqual","value":"5"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseMeetPeriodStart, rule.Field)
				assert.Equal(t, FieldTypeNumber, rule.Type)
				assert.Equal(t, FieldFilterGreaterOrEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "5", *rule.TextValue)
			},
		},
		{
			name:  "valid evening period",
			input: `{"field":"meet_times.period_end","type":"number","filter":"equal","value":"E2"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseMeetPeriodEnd, rule.Field)
				assert.Equal(t, FieldTypeNumber, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "E2", *rule.TextValue)
			},
		},
		{
			name:      "invalid number field",
			input:     `{"field":"course_name","type":"number","filter":"equal","value":10}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid number filter",
			input:     `{"field":"course_level","type":"number","filter":"contains","value":10}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid period",
			input:     `{"field":"meet_times.period_start","type":"number","filter":"equal","value":"INVALID"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "period has wrong type",
			input:     `{"field":"meet_times.period_start","type":"number","filter":"equal","value":5}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "number has wrong type",
			input:     `{"field":"course_level","type":"number","filter":"equal","value":"3000"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:  "valid time",
			input: `{"field":"meet_times.time_start","type":"time","filter":"greater","value":"08:30"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseMeetTimeStart, rule.Field)
				assert.Equal(t, FieldTypeTime, rule.Type)
				assert.Equal(t, FieldFilterGreater, rule.Filter)

				require.NotNil(t, rule.TimeValue)
				assert.Equal(t, 8, rule.TimeValue.Hour)
				assert.Equal(t, 30, rule.TimeValue.Minute)
			},
		},
		{
			name:  "valid late time",
			input: `{"field":"meet_times.time_end","type":"time","filter":"less","value":"23:59"}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseMeetTimeEnd, rule.Field)
				assert.Equal(t, FieldTypeTime, rule.Type)
				assert.Equal(t, FieldFilterLess, rule.Filter)

				require.NotNil(t, rule.TimeValue)
				assert.Equal(t, 23, rule.TimeValue.Hour)
				assert.Equal(t, 59, rule.TimeValue.Minute)
			},
		},
		{
			name:  "valid midnight",
			input: `{"field":"meet_times.time_start","type":"time","filter":"equal","value":"00:00"}`,
			check: func(t *testing.T, rule Rule) {
				require.NotNil(t, rule.TimeValue)
				assert.Equal(t, 0, rule.TimeValue.Hour)
				assert.Equal(t, 0, rule.TimeValue.Minute)
			},
		},
		{
			name:      "invalid time field",
			input:     `{"field":"course_name","type":"time","filter":"equal","value":"08:30"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid time filter",
			input:     `{"field":"meet_times.time_start","type":"time","filter":"contains","value":"08:30"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid time",
			input:     `{"field":"meet_times.time_start","type":"time","filter":"equal","value":"25:30"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid minute",
			input:     `{"field":"meet_times.time_start","type":"time","filter":"equal","value":"12:60"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid time format",
			input:     `{"field":"meet_times.time_start","type":"time","filter":"equal","value":"invalid"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "time value has wrong type",
			input:     `{"field":"meet_times.time_start","type":"time","filter":"equal","value":830}`,
			check:     nil,
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
			input: `{"field":"course_is_ai","type":"boolean","filter":"notEqual","value":false}`,
			check: func(t *testing.T, rule Rule) {
				assert.Equal(t, FieldCourseIsAI, rule.Field)
				assert.Equal(t, FieldTypeBoolean, rule.Type)
				assert.Equal(t, FieldFilterNotEqual, rule.Filter)

				require.NotNil(t, rule.BooleanValue)
				assert.False(t, *rule.BooleanValue)
			},
		},
		{
			name:      "invalid boolean field",
			input:     `{"field":"course_name","type":"boolean","filter":"equal","value":true}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid boolean filter",
			input:     `{"field":"course_is_lab","type":"boolean","filter":"greater","value":true}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "boolean has wrong type",
			input:     `{"field":"course_is_lab","type":"boolean","filter":"equal","value":"true"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "missing value",
			input:     `{"field":"course_name","type":"text","filter":"equal"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "invalid field type",
			input:     `{"field":"course_name","type":"invalid","filter":"equal","value":"test"}`,
			check:     nil,
			wantError: true,
		},
		{
			name:      "malformed JSON",
			input:     `{"field":"course_name","type":"text","filter":"equal","value":}`,
			check:     nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule Rule

			err := json.Unmarshal([]byte(tt.input), &rule)

			if tt.wantError {
				assert.Error(t, err)
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
						"field": "course_name",
						"type": "text",
						"filter": "equal",
						"value": "Computer Science"
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 1)

				rule, ok := filter.Rules[0].(Rule)
				require.True(t, ok)

				assert.Equal(t, FieldCourseName, rule.Field)
				assert.Equal(t, FieldTypeText, rule.Type)
				assert.Equal(t, FieldFilterEqual, rule.Filter)

				require.NotNil(t, rule.TextValue)
				assert.Equal(t, "Computer Science", *rule.TextValue)
			},
		},
		{
			name: "multiple rules with and",
			input: `{
				"glue": "and",
				"rules": [
					{
						"field": "course_level",
						"type": "number",
						"filter": "greaterOrEqual",
						"value": 3000
					},
					{
						"field": "course_is_lab",
						"type": "boolean",
						"filter": "equal",
						"value": false
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 2)

				numberRule, ok := filter.Rules[0].(Rule)
				require.True(t, ok)

				assert.Equal(t, FieldCourseLevel, numberRule.Field)
				require.NotNil(t, numberRule.NumberValue)
				assert.Equal(t, float64(3000), *numberRule.NumberValue)

				booleanRule, ok := filter.Rules[1].(Rule)
				require.True(t, ok)

				assert.Equal(t, FieldCourseIsLab, booleanRule.Field)
				require.NotNil(t, booleanRule.BooleanValue)
				assert.False(t, *booleanRule.BooleanValue)
			},
		},
		{
			name: "or rules",
			input: `{
				"glue": "or",
				"rules": [
					{
						"field": "course_code",
						"type": "text",
						"filter": "equal",
						"value": "COP3502"
					},
					{
						"field": "course_code",
						"type": "text",
						"filter": "equal",
						"value": "CEN3031"
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueOr, filter.Glue)
				require.Len(t, filter.Rules, 2)

				for _, node := range filter.Rules {
					_, ok := node.(Rule)
					assert.True(t, ok)
				}
			},
		},
		{
			name: "nested filter",
			input: `{
				"glue": "and",
				"rules": [
					{
						"field": "course_level",
						"type": "number",
						"filter": "greaterOrEqual",
						"value": 3000
					},
					{
						"glue": "or",
						"rules": [
							{
								"field": "course_is_lab",
								"type": "boolean",
								"filter": "equal",
								"value": true
							},
							{
								"field": "course_is_ai",
								"type": "boolean",
								"filter": "equal",
								"value": true
							}
						]
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 2)

				numberRule, ok := filter.Rules[0].(Rule)
				require.True(t, ok)

				assert.Equal(t, FieldCourseLevel, numberRule.Field)

				nested, ok := filter.Rules[1].(Filter)
				require.True(t, ok)

				assert.Equal(t, GlueOr, nested.Glue)
				require.Len(t, nested.Rules, 2)

				for _, node := range nested.Rules {
					_, ok := node.(Rule)
					assert.True(t, ok)
				}
			},
		},
		{
			name: "nested filter with time rule",
			input: `{
				"glue": "and",
				"rules": [
					{
						"field": "meet_times.time_start",
						"type": "time",
						"filter": "greaterOrEqual",
						"value": "08:30"
					},
					{
						"field": "course_is_lab",
						"type": "boolean",
						"filter": "equal",
						"value": false
					}
				]
			}`,
			check: func(t *testing.T, filter Filter) {
				assert.Equal(t, GlueAnd, filter.Glue)
				require.Len(t, filter.Rules, 2)

				timeRule, ok := filter.Rules[0].(Rule)
				require.True(t, ok)

				require.NotNil(t, timeRule.TimeValue)
				assert.Equal(t, 8, timeRule.TimeValue.Hour)
				assert.Equal(t, 30, timeRule.TimeValue.Minute)

				booleanRule, ok := filter.Rules[1].(Rule)
				require.True(t, ok)

				require.NotNil(t, booleanRule.BooleanValue)
				assert.False(t, *booleanRule.BooleanValue)
			},
		},
		{
			name: "invalid glue",
			input: `{
				"glue": "invalid",
				"rules": [
					{
						"field": "course_name",
						"type": "text",
						"filter": "equal",
						"value": "test"
					}
				]
			}`,
			check:     nil,
			wantError: true,
		},
		{
			name: "empty rules",
			input: `{
				"glue": "and",
				"rules": []
			}`,
			check:     nil,
			wantError: true,
		},
		{
			name: "invalid filter node",
			input: `{
				"glue": "and",
				"rules": [
					{
						"invalid": "node"
					}
				]
			}`,
			check:     nil,
			wantError: true,
		},
		{
			name: "invalid nested filter",
			input: `{
				"glue": "and",
				"rules": [
					{
						"glue": "invalid",
						"rules": []
					}
				]
			}`,
			check:     nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter Filter

			err := json.Unmarshal([]byte(tt.input), &filter)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.check != nil {
				tt.check(t, filter)
			}
		})
	}
}
