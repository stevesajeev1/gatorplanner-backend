package optimizer

import (
	"context"

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
	result := make([]schedulerv1.Day, 0, len(days))

	for _, day := range days {
		switch day {
		case sqlc.MeetDayTypeM:
			result = append(result, schedulerv1.Day_DAY_MONDAY)
		case sqlc.MeetDayTypeT:
			result = append(result, schedulerv1.Day_DAY_TUESDAY)
		case sqlc.MeetDayTypeW:
			result = append(result, schedulerv1.Day_DAY_WEDNESDAY)
		case sqlc.MeetDayTypeR:
			result = append(result, schedulerv1.Day_DAY_THURSDAY)
		case sqlc.MeetDayTypeF:
			result = append(result, schedulerv1.Day_DAY_FRIDAY)
		case sqlc.MeetDayTypeS:
			result = append(result, schedulerv1.Day_DAY_SATURDAY)
		case sqlc.MeetDayTypeU:
			result = append(result, schedulerv1.Day_DAY_SUNDAY)
		}
	}

	return result
}

func timeToMinutes(t pgtype.Time) int32 {
	return int32(t.Microseconds / (60 * 1_000_000))
}

func buildSchedulerRequest(
	classes []sqlc.ListClassesForSchedulerRow,
) *schedulerv1.GenerateSchedulesRequest {
	courses := []*schedulerv1.Course{}

	var currentCourse *schedulerv1.Course
	var currentClass *schedulerv1.Class

	for _, class := range classes {
		// New course
		if currentCourse == nil || currentCourse.CourseId != class.CourseID {
			currentCourse = &schedulerv1.Course{
				CourseId: class.CourseID,
			}

			courses = append(courses, currentCourse)
			currentClass = nil
		}

		// New class
		if currentClass == nil || currentClass.ClassId != class.ClassID.String() {
			currentClass = &schedulerv1.Class{
				ClassId: class.ClassID.String(),
			}

			currentCourse.Classes = append(
				currentCourse.Classes,
				currentClass,
			)
		}

		currentClass.Meetings = append(
			currentClass.Meetings,
			&schedulerv1.Meeting{
				Days:        convertDays(class.Days),
				StartMinute: timeToMinutes(class.TimeBegin),
				EndMinute:   timeToMinutes(class.TimeEnd),
			},
		)
	}

	return &schedulerv1.GenerateSchedulesRequest{
		Courses: courses,
	}
}

func (r *SchedulesOptimizerRepository) GenerateSchedules(ctx context.Context, classes []sqlc.ListClassesForSchedulerRow) (*schedulerv1.GenerateSchedulesResponse, error) {
	request := buildSchedulerRequest(classes)
	return r.client.GenerateSchedules(ctx, request)
}
