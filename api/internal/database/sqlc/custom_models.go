package sqlc

import (
	"encoding/json"
	"fmt"
)

var ValidMeetTypes = []ClassMeetType{ClassMeetTypePrimarilyClassroom, ClassMeetTypeHybrid, ClassMeetTypeOnline8099, ClassMeetTypeOnline100}
var ValidGenEds = []GenEd{GenEdBiologicalScience, GenEdPhysicalScience, GenEdSocialScience, GenEdMathematics, GenEdComposition, GenEdHumanities, GenEdInternational}
var ValidQuests = []Quest{QuestQuest1, QuestQuest2, QuestQuest3, QuestQuest4}
var ValidMeetDays = []MeetDayType{MeetDayTypeM, MeetDayTypeT, MeetDayTypeW, MeetDayTypeR, MeetDayTypeF, MeetDayTypeS, MeetDayTypeU}

type CustomCourseCredits struct {
	Gte float64 `json:"gte"`
	Lte float64 `json:"lte"`
}

type CustomMeetTime struct {
	Days        []MeetDayType `json:"days"`
	TimeStart   *int32        `json:"time_start"`
	TimeEnd     *int32        `json:"time_end"`
	PeriodStart *int32        `json:"period_start"`
	PeriodEnd   *int32        `json:"period_end"`
	Building    *string       `json:"building"`
}

type CustomInstructor struct {
	Name       string   `json:"name"`
	Rating     *float64 `json:"rating"`
	Difficulty *float64 `json:"difficulty"`
	TakeAgain  *float64 `json:"take_again"`
}

type TypedSearchClassesRow struct {
	SearchClassesRow
	ID            string              `json:"id"`
	CourseCredits CustomCourseCredits `json:"course_credits"`
	CourseGenEds  []GenEd             `json:"course_gen_eds"`
	CourseQuest   *Quest              `json:"course_quest"`
	MeetTimes     []CustomMeetTime    `json:"meet_times"`
	Instructors   []CustomInstructor  `json:"instructors"`
}

func (r *TypedSearchClassesRow) UnmarshalJSON(data []byte) error {
	var custom struct {
		CourseCredits CustomCourseCredits `json:"course_credits"`
		CourseGenEds  []GenEd             `json:"course_gen_eds"`
		CourseQuest   *Quest              `json:"course_quest"`
		MeetTimes     []CustomMeetTime    `json:"meet_times"`
		Instructors   []CustomInstructor  `json:"instructors"`
	}

	if err := json.Unmarshal(data, &custom); err != nil {
		return fmt.Errorf("unmarshal custom fields: %w", err)
	}

	var base map[string]json.RawMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return fmt.Errorf("unmarshal base fields: %w", err)
	}

	delete(base, "course_credits")
	delete(base, "course_gen_eds")
	delete(base, "course_quest")
	delete(base, "meet_times")
	delete(base, "instructors")

	baseData, err := json.Marshal(base)
	if err != nil {
		return fmt.Errorf("marshal base fields: %w", err)
	}

	if err := json.Unmarshal(baseData, &r.SearchClassesRow); err != nil {
		return fmt.Errorf("unmarshal SearchClassesRow: %w", err)
	}

	r.CourseCredits = custom.CourseCredits
	r.CourseGenEds = custom.CourseGenEds
	r.CourseQuest = custom.CourseQuest
	r.MeetTimes = custom.MeetTimes
	r.Instructors = custom.Instructors

	return nil
}
