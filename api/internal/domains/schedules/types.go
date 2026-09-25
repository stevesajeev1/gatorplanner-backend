package schedules

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	schedulerv1 "github.com/stevesajeev1/gatorplanner-backend/generated/scheduler/v1"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

type SortBy string

var SortByMap = map[schedulerv1.SortBy]SortBy{
	schedulerv1.SortBy_SORT_BY_EARLIEST_START:    "earliest_start",
	schedulerv1.SortBy_SORT_BY_LATEST_START:      "latest_start",
	schedulerv1.SortBy_SORT_BY_EARLIEST_END:      "earliest_end",
	schedulerv1.SortBy_SORT_BY_LATEST_END:        "latest_end",
	schedulerv1.SortBy_SORT_BY_MOST_COMPACT:      "most_compact",
	schedulerv1.SortBy_SORT_BY_FEWEST_DAYS:       "fewest_days",
	schedulerv1.SortBy_SORT_BY_MOST_BALANCED:     "most_balanced",
	schedulerv1.SortBy_SORT_BY_INSTRUCTOR_RATING: "instructor_rating",
}

func (SortBy) Schema(r huma.Registry) *huma.Schema {
	values := make([]any, 0, len(SortByMap))
	for _, value := range SortByMap {
		values = append(values, string(value))
	}

	return &huma.Schema{
		Type: huma.TypeString,
		Enum: values,
	}
}

type ClassChoice struct {
	CourseID uuid.UUID   `json:"course_id"`
	ClassIDs []uuid.UUID `json:"class_ids" nullable:"false" minItems:"1" uniqueItems:"true"`
}

type SelectedClass struct {
	CourseID uuid.UUID `json:"course_id"`
	ClassID  uuid.UUID `json:"class_id"`
}

type GenerateSchedulesRequest struct {
	ClassChoices    []ClassChoice      `json:"class_choices" nullable:"false"`
	SortBy          *SortBy            `json:"sort_by,omitempty"`
	DayRestrictions []sqlc.MeetDayType `json:"day_restrictions,omitempty" nullable:"false" default:"[]"`
}

type Schedule struct {
	Classes []SelectedClass `json:"classes" nullable:"false"`
}

type GenerateSchedulesResponse struct {
	Body []Schedule `nullable:"false"`
}
