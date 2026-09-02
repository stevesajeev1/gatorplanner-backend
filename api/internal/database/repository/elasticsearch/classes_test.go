package elasticsearch

import (
	"encoding/json"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/classes/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearchQuery(t *testing.T) {
	query := buildSearchQuery("algorithms")

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	boolQuery, ok := got["bool"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, float64(1), boolQuery["minimum_should_match"])

	should, ok := boolQuery["should"].([]any)
	require.True(t, ok)
	require.Len(t, should, 5)

	assert.Contains(t, string(data), `"multi_match"`)
	assert.Contains(t, string(data), `"instructors"`)
	assert.Contains(t, string(data), `"meet_times"`)
	assert.Contains(t, string(data), `"course_code.keyword"`)
	assert.Contains(t, string(data), `"course_department.keyword"`)
}

func TestBuildTermFilter(t *testing.T) {
	query := esdsl.NewBoolQuery().
		Filter(
			esdsl.NewTermQuery(
				"term_id",
				esdsl.NewFieldValue().Int64(2268),
			),
		)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	boolQuery := got["bool"].(map[string]any)

	filters := boolQuery["filter"].([]any)
	require.Len(t, filters, 1)

	term := filters[0].(map[string]any)
	termBody := term["term"].(map[string]any)
	termID := termBody["term_id"].(map[string]any)

	assert.Equal(t, float64(2268), termID["value"])
}

func TestBuildFilterQuery_And(t *testing.T) {
	value := true

	filter := search.Filter{
		Glue: search.GlueAnd,
		Rules: []search.FilterNode{
			search.Rule{
				Field:        search.FieldCourseIsAI,
				Type:         search.FieldTypeBoolean,
				Filter:       search.FieldFilterEqual,
				BooleanValue: &value,
			},
		},
	}

	query := buildFilterQuery(filter)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.Contains(t, string(data), `"course_is_ai"`)
	assert.Contains(t, string(data), `"value":true`)
}

func TestBuildFilterQuery_Or(t *testing.T) {
	value1 := true
	value2 := false

	filter := search.Filter{
		Glue: search.GlueOr,
		Rules: []search.FilterNode{
			search.Rule{
				Field:        search.FieldCourseIsAI,
				Type:         search.FieldTypeBoolean,
				Filter:       search.FieldFilterEqual,
				BooleanValue: &value1,
			},
			search.Rule{
				Field:        search.FieldCourseIsHonors,
				Type:         search.FieldTypeBoolean,
				Filter:       search.FieldFilterEqual,
				BooleanValue: &value2,
			},
		},
	}

	query := buildFilterQuery(filter)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.Contains(t, string(data), `"course_is_ai"`)
	assert.Contains(t, string(data), `"course_is_honors"`)
	assert.Contains(t, string(data), `"minimum_should_match":1`)
}

func TestBuildTextRuleQuery_Equal(t *testing.T) {
	value := "COP3502"

	rule := search.Rule{
		Field:     search.FieldCourseCode,
		Type:      search.FieldTypeText,
		Filter:    search.FieldFilterEqual,
		TextValue: &value,
	}

	query := buildTextRuleQuery(rule)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"term": {
			"course_code.keyword": {
				"value": "COP3502"
			}
		}
	}`, string(data))
}

func TestBuildTextRuleQuery_NotEqual(t *testing.T) {
	value := "COP3502"

	rule := search.Rule{
		Field:     search.FieldCourseCode,
		Type:      search.FieldTypeText,
		Filter:    search.FieldFilterNotEqual,
		TextValue: &value,
	}

	query := buildTextRuleQuery(rule)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.Contains(t, string(data), `"must_not"`)
	assert.Contains(t, string(data), `"course_code.keyword"`)
	assert.Contains(t, string(data), `"COP3502"`)
}

func TestBuildBooleanRuleQuery_Equal(t *testing.T) {
	value := true

	rule := search.Rule{
		Field:        search.FieldCourseIsAI,
		Type:         search.FieldTypeBoolean,
		Filter:       search.FieldFilterEqual,
		BooleanValue: &value,
	}

	query := buildBooleanRuleQuery(rule)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"term": {
			"course_is_ai": {
				"value": true
			}
		}
	}`, string(data))
}

func TestBuildBooleanRuleQuery_NotEqual(t *testing.T) {
	value := true

	rule := search.Rule{
		Field:        search.FieldCourseIsAI,
		Type:         search.FieldTypeBoolean,
		Filter:       search.FieldFilterNotEqual,
		BooleanValue: &value,
	}

	query := buildBooleanRuleQuery(rule)

	data, err := query.QueryCaster().MarshalJSON()
	require.NoError(t, err)

	assert.Contains(t, string(data), `"must_not"`)
	assert.Contains(t, string(data), `"course_is_ai"`)
}

func TestBuildScalarNumberRuleQuery(t *testing.T) {
	tests := []struct {
		name   string
		filter search.FieldFilter
		want   string
	}{
		{
			name:   "equal",
			filter: search.FieldFilterEqual,
			want:   `"value":3`,
		},
		{
			name:   "not equal",
			filter: search.FieldFilterNotEqual,
			want:   `"must_not"`,
		},
		{
			name:   "greater",
			filter: search.FieldFilterGreater,
			want:   `"gt":3`,
		},
		{
			name:   "less",
			filter: search.FieldFilterLess,
			want:   `"lt":3`,
		},
		{
			name:   "greater or equal",
			filter: search.FieldFilterGreaterOrEqual,
			want:   `"gte":3`,
		},
		{
			name:   "less or equal",
			filter: search.FieldFilterLessOrEqual,
			want:   `"lte":3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := float64(3)

			rule := search.Rule{
				Field:       search.FieldInstructorRating,
				Type:        search.FieldTypeNumber,
				Filter:      tt.filter,
				NumberValue: &value,
			}

			query := buildScalarNumberRuleQuery(rule)

			data, err := query.QueryCaster().MarshalJSON()
			require.NoError(t, err)

			assert.Contains(t, string(data), tt.want)
		})
	}
}

func TestBuildRangeNumberRuleQuery(t *testing.T) {
	tests := []struct {
		name   string
		filter search.FieldFilter
		want   string
	}{
		{
			name:   "equal",
			filter: search.FieldFilterEqual,
			want:   `"relation":"contains"`,
		},
		{
			name:   "not equal",
			filter: search.FieldFilterNotEqual,
			want:   `"must_not"`,
		},
		{
			name:   "greater",
			filter: search.FieldFilterGreater,
			want:   `"gt":3`,
		},
		{
			name:   "less",
			filter: search.FieldFilterLess,
			want:   `"lt":3`,
		},
		{
			name:   "greater or equal",
			filter: search.FieldFilterGreaterOrEqual,
			want:   `"gte":3`,
		},
		{
			name:   "less or equal",
			filter: search.FieldFilterLessOrEqual,
			want:   `"lte":3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := float64(3)

			rule := search.Rule{
				Field:       search.FieldCourseCredits,
				Type:        search.FieldTypeNumber,
				Filter:      tt.filter,
				NumberValue: &value,
			}

			query := buildRangeNumberRuleQuery(rule)

			data, err := query.QueryCaster().MarshalJSON()
			require.NoError(t, err)

			assert.Contains(t, string(data), tt.want)
		})
	}
}

func TestKeywordField(t *testing.T) {
	tests := []struct {
		field search.Field
		want  string
	}{
		{search.FieldCourseCode, "course_code.keyword"},
		{search.FieldCourseName, "course_name.keyword"},
		{search.FieldCourseDepartment, "course_department.keyword"},
		{search.FieldCourseGenEds, "course_gen_eds.keyword"},
		{search.FieldCourseQuest, "course_quest.keyword"},
		{search.FieldCourseMeetBuilding, "meet_times.building.keyword"},
		{search.FieldInstructorName, "instructors.name.keyword"},
	}

	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			assert.Equal(t, tt.want, keywordField(tt.field))
		})
	}
}

func TestWrapNestedQuery(t *testing.T) {
	tests := []struct {
		name  string
		field search.Field
		path  string
	}{
		{
			name:  "meet days",
			field: search.FieldCourseMeetDays,
			path:  "meet_times",
		},
		{
			name:  "meet time start",
			field: search.FieldCourseMeetTimeStart,
			path:  "meet_times",
		},
		{
			name:  "meet time end",
			field: search.FieldCourseMeetTimeEnd,
			path:  "meet_times",
		},
		{
			name:  "meet period start",
			field: search.FieldCourseMeetPeriodStart,
			path:  "meet_times",
		},
		{
			name:  "meet period end",
			field: search.FieldCourseMeetPeriodEnd,
			path:  "meet_times",
		},
		{
			name:  "building",
			field: search.FieldCourseMeetBuilding,
			path:  "meet_times",
		},
		{
			name:  "instructor name",
			field: search.FieldInstructorName,
			path:  "instructors",
		},
		{
			name:  "instructor rating",
			field: search.FieldInstructorRating,
			path:  "instructors",
		},
		{
			name:  "instructor difficulty",
			field: search.FieldInstructorDifficulty,
			path:  "instructors",
		},
		{
			name:  "instructor take again",
			field: search.FieldInstructorTakeAgain,
			path:  "instructors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := wrapNestedQuery(
				tt.field,
				esdsl.NewTermQuery(
					string(tt.field),
					esdsl.NewFieldValue().String("test"),
				),
			)

			data, err := query.QueryCaster().MarshalJSON()
			require.NoError(t, err)

			assert.Contains(t, string(data), `"nested"`)
			assert.Contains(t, string(data), `"path":"`+tt.path+`"`)
		})
	}
}
