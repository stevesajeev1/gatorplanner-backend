package search

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
	"github.com/stevesajeev1/gatorplanner-backend/internal/util"
)

func RegisterCustomSchemas(r huma.Registry) {
	r.Map()["FilterRuleTextClassMeetType"] = (Rule{}).ClassMeetTypeTextSchema(r)
	r.Map()["FilterRuleTextCourseGenEds"] = (Rule{}).CourseGenEdsTextSchema(r)
	r.Map()["FilterRuleTextCourseQuest"] = (Rule{}).CourseQuestTextSchema(r)
	r.Map()["FilterRuleTextCourseMeetDays"] = (Rule{}).CourseMeetDaysTextSchema(r)
	r.Map()["FilterRuleTextGeneralField"] = (Rule{}).GeneralTextSchema(r)
	r.Map()["FilterRuleText"] = (Rule{}).TextSchema(r)
	r.Map()["FilterRuleNumberPeriod"] = (Rule{}).PeriodNumberSchema(r)
	r.Map()["FilterRuleNumberGeneralField"] = (Rule{}).GeneralNumberSchema(r)
	r.Map()["FilterRuleNumber"] = (Rule{}).NumberSchema(r)
	r.Map()["FilterRuleTime"] = (Rule{}).TimeSchema(r)
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
				Enum: util.AnyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(sqlc.ValidMeetTypes),
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
				Enum: util.AnyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(sqlc.ValidGenEds),
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
				Enum: util.AnyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(sqlc.ValidQuests),
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
				Enum: util.AnyStringSlice(ValidFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(sqlc.ValidMeetDays),
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
				Enum: util.AnyStringSlice(ValidGeneralTextFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeText)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidFilters),
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

func (Rule) PeriodNumberSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: []any{string(FieldCourseMeetPeriodStart), string(FieldCourseMeetPeriodEnd)},
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeNumber)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidNumberFilters),
			},
			"value": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(sqlc.ValidPeriods),
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

func (Rule) GeneralNumberSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidGeneralNumberFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeNumber)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidNumberFilters),
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

func (Rule) NumberSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{
				Ref: "#/components/schemas/FilterRuleNumberPeriod",
			},
			{
				Ref: "#/components/schemas/FilterRuleNumberGeneralField",
			},
		},
	}
}

func (Rule) TimeSchema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: huma.TypeObject,
		Properties: map[string]*huma.Schema{
			"field": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidTimeFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeTime)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidNumberFilters),
			},
			"value": {
				Type:   huma.TypeString,
				Format: "time",
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
				Enum: util.AnyStringSlice(ValidBooleanFields),
			},
			"type": {
				Type: huma.TypeString,
				Enum: []any{string(FieldTypeBoolean)},
			},
			"filter": {
				Type: huma.TypeString,
				Enum: util.AnyStringSlice(ValidFilters),
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
				Ref: "#/components/schemas/FilterRuleTime",
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
				Enum: util.AnyStringSlice([]Glue{GlueAnd, GlueOr}),
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
