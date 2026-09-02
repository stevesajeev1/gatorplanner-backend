package search

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

func RegisterCustomSchemas(r huma.Registry) {
	r.Map()["FilterRuleTextClassMeetType"] = (Rule{}).ClassMeetTypeTextSchema(r)
	r.Map()["FilterRuleTextCourseGenEds"] = (Rule{}).CourseGenEdsTextSchema(r)
	r.Map()["FilterRuleTextCourseQuest"] = (Rule{}).CourseQuestTextSchema(r)
	r.Map()["FilterRuleTextCourseMeetDays"] = (Rule{}).CourseMeetDaysTextSchema(r)
	r.Map()["FilterRuleTextGeneralField"] = (Rule{}).GeneralTextSchema(r)
	r.Map()["FilterRuleText"] = (Rule{}).TextSchema(r)
	r.Map()["FilterRuleNumber"] = (Rule{}).NumberSchema(r)
	r.Map()["FilterRuleBoolean"] = (Rule{}).BooleanSchema(r)
	r.Map()["FilterRule"] = (Rule{}).Schema(r)
	r.Map()["Filter"] = (Filter{}).Schema(r)
}

func (Rule) ClassMeetTypeTextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: []any{string(FieldClassMeetType)},
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: anyStringSlice(sqlc.ValidMeetTypes),
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) CourseGenEdsTextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: []any{string(FieldCourseGenEds)},
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: anyStringSlice(sqlc.ValidGenEds),
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) CourseQuestTextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: []any{string(FieldCourseQuest)},
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: anyStringSlice(sqlc.ValidQuests),
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) CourseMeetDaysTextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: []any{string(FieldCourseMeetDays)},
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: anyStringSlice(sqlc.ValidMeetDays),
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) GeneralTextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidGeneralTextFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) TextSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{
				Ref: "#/components/schemas/FilterRuleTextClassMeetType",
			},
			{
				Ref: "#/components/schemas/FilterRuleTextCourseGenEds",
			},
			{
				Ref: "#/components/schemas/FilterRuleTextCourseQuest",
			},
			{
				Ref: "#/components/schemas/FilterRuleTextCourseMeetDays",
			},
			{
				Ref: "#/components/schemas/FilterRuleTextGeneralField",
			},
		},
	}
}

func (Rule) NumberSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidNumberFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeNumber)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidNumberFilters),
			},
			"value": {
				Type: huma.TypeNumber,
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) BooleanSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidBooleanFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeBoolean)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: anyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeBoolean,
			},
		},
		Required: []string{
			"field",
			"type",
			"filter",
			"value",
		},
	}
}

func (Rule) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{
				Ref: "#/components/schemas/FilterRuleText",
			},
			{
				Ref: "#/components/schemas/FilterRuleNumber",
			},
			{
				Ref: "#/components/schemas/FilterRuleBoolean",
			},
		},
	}
}

func (Filter) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"glue": {
				Type: huma.TypeString,
				Enum: anyStringSlice([]Glue{GlueAnd, GlueOr}),
			},
			"rules": {
				Type:     huma.TypeArray,
				MinItems: new(1),
				Items: &huma.Schema{
					OneOf: []*huma.Schema{
						{
							Ref: "#/components/schemas/FilterRule",
						},
						{
							Ref: "#/components/schemas/Filter",
						},
					},
				},
			},
		},
		Required: []string{"glue", "rules"},
	}
}

func anyStringSlice[T ~string](slice []T) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = string(v)
	}
	return result
}
