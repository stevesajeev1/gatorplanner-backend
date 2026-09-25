package optimizer

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	schedulerv1 "github.com/stevesajeev1/gatorplanner-backend/generated/scheduler/v1"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
)

type SchedulesOptimizerRepository struct {
	client schedulerv1.SchedulerServiceClient
}

func NewSchedulesOptimizerRepository(o *dependencies.Optimizer) *SchedulesOptimizerRepository {
	client := schedulerv1.NewSchedulerServiceClient(o.ClientConn)
	return &SchedulesOptimizerRepository{client: client}
}

func convertDays(days []sqlc.MeetDayType) []schedulerv1.Day {
	result := make([]schedulerv1.Day, len(days))

	for i, day := range days {
		switch day {
		case sqlc.MeetDayTypeM:
			result[i] = schedulerv1.Day_DAY_MONDAY
		case sqlc.MeetDayTypeT:
			result[i] = schedulerv1.Day_DAY_TUESDAY
		case sqlc.MeetDayTypeW:
			result[i] = schedulerv1.Day_DAY_WEDNESDAY
		case sqlc.MeetDayTypeR:
			result[i] = schedulerv1.Day_DAY_THURSDAY
		case sqlc.MeetDayTypeF:
			result[i] = schedulerv1.Day_DAY_FRIDAY
		case sqlc.MeetDayTypeS:
			result[i] = schedulerv1.Day_DAY_SATURDAY
		case sqlc.MeetDayTypeU:
			result[i] = schedulerv1.Day_DAY_SUNDAY
		}
	}

	return result
}

func timeToMinutes(t pgtype.Time) int32 {
	return int32(t.Microseconds / (60 * 1_000_000))
}

func buildSchedulerRequest(
	classes []sqlc.TypedListClassesForSchedulerRow,
) (*schedulerv1.GenerateSchedulesRequest, error) {
	courses := []*schedulerv1.Course{}

	var currentCourse *schedulerv1.Course = nil
	var currentClass *schedulerv1.Class = nil

	for _, class := range classes {
		// New course
		if currentCourse == nil || currentCourse.CourseId != class.CourseID.String() {
			currentCourse = &schedulerv1.Course{
				CourseId: class.CourseID.String(),
			}

			courses = append(courses, currentCourse)
			currentClass = nil
		}

		// New class
		if currentClass == nil || currentClass.ClassId != class.ClassID.String() {
			currentClass = &schedulerv1.Class{
				ClassId: class.ClassID.String(),
			}

			// Add instructor rating if available
			if class.AvgInstructorRating.Valid {
				avgInstructorRating, err := class.AvgInstructorRating.Float64Value()
				if err != nil {
					return nil, fmt.Errorf("convert avg instructor rating for class %s: %w", class.ClassID, err)
				}

				currentClass.AvgInstructorRating = &avgInstructorRating.Float64
			}

			currentCourse.Classes = append(
				currentCourse.Classes,
				currentClass,
			)
		}

		// No meeting (class could be online)
		if class.Days == nil || !class.TimeBegin.Valid || !class.TimeEnd.Valid {
			continue
		}

		meeting := &schedulerv1.Meeting{
			Days:        convertDays(class.Days),
			StartMinute: timeToMinutes(class.TimeBegin),
			EndMinute:   timeToMinutes(class.TimeEnd),
		}

		// Add building location if available
		if class.BuildingLatitude.Valid && class.BuildingLongitude.Valid {
			buildingLatitude, err := class.BuildingLatitude.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("convert building latitude for class %s: %w", class.ClassID, err)
			}
			buildingLongitude, err := class.BuildingLongitude.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("convert building longitude for class %s: %w", class.ClassID, err)
			}

			meeting.BuildingLatitude = &buildingLatitude.Float64
			meeting.BuildingLongitude = &buildingLongitude.Float64
		}

		currentClass.Meetings = append(
			currentClass.Meetings,
			meeting,
		)
	}

	return &schedulerv1.GenerateSchedulesRequest{
		Courses: courses,
	}, nil
}

func (r *SchedulesOptimizerRepository) GenerateSchedules(ctx context.Context, classes []sqlc.TypedListClassesForSchedulerRow, sortBy *schedulerv1.SortBy, dayRestrictions []sqlc.MeetDayType) (*schedulerv1.GenerateSchedulesResponse, error) {
	request, err := buildSchedulerRequest(classes)
	if err != nil {
		return nil, err
	}

	if sortBy != nil {
		request.SortBy = sortBy
	}

	if len(dayRestrictions) > 0 {
		request.DayRestrictions = convertDays(dayRestrictions)
	}

	return r.client.GenerateSchedules(ctx, request)
}
