package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

// Inspired from https://docs.svar.dev/svelte/filter/api/properties/filterbuilder-value/

type FilterNode interface {
	isFilterNode()
}

type Glue string

const (
	GlueAnd Glue = "and"
	GlueOr  Glue = "or"
)

type Filter struct {
	Glue  Glue         `json:"glue"`
	Rules []FilterNode `json:"rules"`
}

type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
)

type Field string

// Text Fields
const (
	FieldClassMeetType      Field = "meet_type"
	FieldCourseCode         Field = "course_code"
	FieldCourseCodePrefix   Field = "course_code_prefix"
	FieldCourseName         Field = "course_name"
	FieldCourseDepartment   Field = "course_department"
	FieldCourseGenEds       Field = "course_gen_eds"
	FieldCourseQuest        Field = "course_quest"
	FieldCourseMeetDays     Field = "meet_times.days"
	FieldCourseMeetBuilding Field = "meet_times.building"
	FieldInstructorName     Field = "instructors.name"
)

var ValidGeneralTextFields = []Field{FieldCourseCode, FieldCourseCodePrefix, FieldCourseName, FieldCourseDepartment, FieldCourseMeetBuilding, FieldInstructorName}
var ValidCheckedTextFields = []Field{FieldClassMeetType, FieldCourseGenEds, FieldCourseQuest, FieldCourseMeetDays}

// Number Fields
const (
	FieldClassNumber           Field = "number"
	FieldCourseLevel           Field = "course_level"
	FieldCourseCredits         Field = "course_credits"
	FieldCourseWords           Field = "course_words"
	FieldCourseMeetTimeStart   Field = "meet_times.time_start"
	FieldCourseMeetTimeEnd     Field = "meet_times.time_end"
	FieldCourseMeetPeriodStart Field = "meet_times.period_start"
	FieldCourseMeetPeriodEnd   Field = "meet_times.period_end"
	FieldInstructorRating      Field = "instructors.rating"
	FieldInstructorDifficulty  Field = "instructors.difficulty"
	FieldInstructorTakeAgain   Field = "instructors.take_again"
)

var ValidNumberFields = []Field{FieldClassNumber, FieldCourseLevel, FieldCourseCredits, FieldCourseWords, FieldCourseMeetTimeStart, FieldCourseMeetTimeEnd, FieldCourseMeetPeriodStart, FieldCourseMeetPeriodEnd, FieldInstructorRating, FieldInstructorDifficulty, FieldInstructorTakeAgain}

// Boolean Fields
const (
	FieldCourseIsLab    Field = "course_is_lab"
	FieldCourseIsAI     Field = "course_is_ai"
	FieldCourseIsHonors Field = "course_is_honors"
)

var ValidBooleanFields = []Field{FieldCourseIsLab, FieldCourseIsAI, FieldCourseIsHonors}

type FieldFilter string

const (
	// SHARED
	FieldFilterEqual    FieldFilter = "equal"
	FieldFilterNotEqual FieldFilter = "notEqual"

	// NUMBER
	FieldFilterGreater        FieldFilter = "greater"
	FieldFilterLess           FieldFilter = "less"
	FieldFilterGreaterOrEqual FieldFilter = "greaterOrEqual"
	FieldFilterLessOrEqual    FieldFilter = "lessOrEqual"
)

var ValidFilters = []FieldFilter{FieldFilterEqual, FieldFilterNotEqual}
var ValidNumberFilters = []FieldFilter{FieldFilterEqual, FieldFilterNotEqual, FieldFilterGreater, FieldFilterLess, FieldFilterGreaterOrEqual, FieldFilterLessOrEqual}

type Rule struct {
	Field  Field       `json:"field"`
	Type   FieldType   `json:"type"`
	Filter FieldFilter `json:"filter"`

	TextValue    *string  `json:"-"`
	NumberValue  *float64 `json:"-"`
	BooleanValue *bool    `json:"-"`
}

func (Filter) isFilterNode() {}
func (Rule) isFilterNode()   {}

func (r *Rule) UnmarshalJSON(data []byte) error {
	var raw struct {
		Field  Field            `json:"field"`
		Type   FieldType        `json:"type"`
		Filter FieldFilter      `json:"filter"`
		Value  *json.RawMessage `json:"value,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Value == nil {
		return errors.New("value is required")
	}

	r.Field = raw.Field
	r.Type = raw.Type
	r.Filter = raw.Filter

	switch r.Type {
	case FieldTypeText:
		if !slices.Contains(ValidGeneralTextFields, r.Field) && !slices.Contains(ValidCheckedTextFields, r.Field) {
			return fmt.Errorf("invalid text field %q", r.Field)
		}

		if !slices.Contains(ValidFilters, r.Filter) {
			return fmt.Errorf("invalid text filter %q", r.Filter)
		}

		if err := json.Unmarshal(*raw.Value, &r.TextValue); err != nil {
			return fmt.Errorf("invalid text value: %w", err)
		}

		switch r.Field {
		case FieldClassMeetType:
			if !slices.Contains(sqlc.ValidMeetTypes, sqlc.ClassMeetType(*r.TextValue)) {
				return fmt.Errorf("invalid text value: %q", *r.TextValue)
			}
		case FieldCourseGenEds:
			if !slices.Contains(sqlc.ValidGenEds, sqlc.GenEd(*r.TextValue)) {
				return fmt.Errorf("invalid text value: %q", *r.TextValue)
			}
		case FieldCourseQuest:
			if !slices.Contains(sqlc.ValidQuests, sqlc.Quest(*r.TextValue)) {
				return fmt.Errorf("invalid text value: %q", *r.TextValue)
			}
		case FieldCourseMeetDays:
			if !slices.Contains(sqlc.ValidMeetDays, sqlc.MeetDayType(*r.TextValue)) {
				return fmt.Errorf("invalid text value: %q", *r.TextValue)
			}
		}
	case FieldTypeNumber:
		if !slices.Contains(ValidNumberFields, r.Field) {
			return fmt.Errorf("invalid number field %q", r.Field)
		}

		if !slices.Contains(ValidNumberFilters, r.Filter) {
			return fmt.Errorf("invalid number filter %q", r.Filter)
		}

		if err := json.Unmarshal(*raw.Value, &r.NumberValue); err != nil {
			return fmt.Errorf("invalid number value: %w", err)
		}
	case FieldTypeBoolean:
		if !slices.Contains(ValidBooleanFields, r.Field) {
			return fmt.Errorf("invalid boolean field %q", r.Field)
		}

		if !slices.Contains(ValidFilters, r.Filter) {
			return fmt.Errorf("invalid boolean filter %q", r.Filter)
		}

		if err := json.Unmarshal(*raw.Value, &r.BooleanValue); err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
	default:
		return fmt.Errorf("invalid field type %q", r.Type)
	}

	return nil
}

func (f *Filter) UnmarshalJSON(data []byte) error {
	var raw struct {
		Glue  Glue              `json:"glue"`
		Rules []json.RawMessage `json:"rules"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw.Glue {
	case GlueAnd, GlueOr:
		// Valid glue
	default:
		return fmt.Errorf("invalid glue %q", raw.Glue)
	}

	if len(raw.Rules) == 0 {
		return fmt.Errorf("filter node must have nonempty rules")
	}

	f.Glue = raw.Glue
	f.Rules = make([]FilterNode, 0, len(raw.Rules))

	for _, data := range raw.Rules {
		var discriminator struct {
			Glue  *Glue  `json:"glue"`
			Field *Field `json:"field"`
		}

		if err := json.Unmarshal(data, &discriminator); err != nil {
			return err
		}

		switch {
		case discriminator.Glue != nil:
			var filter Filter

			if err := json.Unmarshal(data, &filter); err != nil {
				return err
			}

			f.Rules = append(f.Rules, filter)

		case discriminator.Field != nil:
			var rule Rule

			if err := json.Unmarshal(data, &rule); err != nil {
				return err
			}

			f.Rules = append(f.Rules, rule)

		default:
			return fmt.Errorf("invalid filter node")
		}
	}

	return nil
}
