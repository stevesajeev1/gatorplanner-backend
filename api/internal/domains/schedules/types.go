package schedules

import (
	"fmt"

	"github.com/google/uuid"
)

type FixedCourse struct {
	CourseID int32     `json:"course_id"`
	ClassID  uuid.UUID `json:"class_id"`
}

type CourseChoice struct {
	CourseID int32       `json:"course_id"`
	ClassIDs []uuid.UUID `json:"class_ids"`
}

type GenerateSchedulesRequest struct {
	FixedCourses  []FixedCourse  `json:"fixed_courses"`
	CourseChoices []CourseChoice `json:"course_choices"`
}

func (r *GenerateSchedulesRequest) Validate() error {
	fixedCoursesSet := make(map[int32]struct{}, len(r.FixedCourses))
	for _, course := range r.FixedCourses {
		fixedCoursesSet[course.CourseID] = struct{}{}
	}

	for _, course := range r.CourseChoices {
		if _, ok := fixedCoursesSet[course.CourseID]; ok {
			return fmt.Errorf(
				"%w: course %d can not be both fixed and have choices",
				ErrInvalidCourses,
				course.CourseID,
			)
		}

		if len(course.ClassIDs) == 0 {
			return ErrCourseMustProvideChoices
		}
	}

	return nil
}

type Schedule struct {
	Courses []FixedCourse `json:"courses"`
}

type GenerateSchedulesResponse struct {
	Body []Schedule
}
