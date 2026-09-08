package sqlc

import (
	"encoding/json"
	"fmt"
	"time"
)

var ValidMeetTypes = []ClassMeetType{ClassMeetTypePrimarilyClassroom, ClassMeetTypeHybrid, ClassMeetTypeOnline8099, ClassMeetTypeOnline100}
var ValidGenEds = []GenEd{GenEdBiologicalScience, GenEdPhysicalScience, GenEdSocialScience, GenEdMathematics, GenEdComposition, GenEdHumanities, GenEdInternational}
var ValidQuests = []Quest{QuestQuest1, QuestQuest2, QuestQuest3, QuestQuest4}
var ValidMeetDays = []MeetDayType{MeetDayTypeM, MeetDayTypeT, MeetDayTypeW, MeetDayTypeR, MeetDayTypeF, MeetDayTypeS, MeetDayTypeU}
var ValidPeriods = []Period{Period1, Period2, Period3, Period4, Period5, Period6, Period7, Period8, Period9, Period10, Period11, PeriodE1, PeriodE2, PeriodE3}

type RawListCoursesByIDRows []ListCoursesByIDRow

type TypedListCoursesByIDRow struct {
	ListCoursesByIDRow
	CreditsMin float64 `json:"credits_min"`
	CreditsMax float64 `json:"credits_max"`
	GenEds     []GenEd `json:"gen_eds"`
	Quest      *Quest  `json:"quest"`
}

func (r RawListCoursesByIDRows) Typed() ([]TypedListCoursesByIDRow, error) {
	typedCourses := make([]TypedListCoursesByIDRow, len(r))
	for i, rawCourse := range r {
		creditsMin, err := rawCourse.CreditsMin.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("convert credits min for course %s: %w", rawCourse.Code, err)
		}
		creditsMax, err := rawCourse.CreditsMax.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("convert credits max for course %s: %w", rawCourse.Code, err)
		}

		var genEds []GenEd
		if err := json.Unmarshal(rawCourse.GenEds, &genEds); err != nil {
			return nil, fmt.Errorf("unmarshal gen eds for course %s: %w", rawCourse.Code, err)
		}

		var quest *Quest = nil
		if rawCourse.Quest.Valid {
			quest = &rawCourse.Quest.Quest
		}

		typedCourses[i] = TypedListCoursesByIDRow{
			ListCoursesByIDRow: rawCourse,
			CreditsMin:         creditsMin.Float64,
			CreditsMax:         creditsMax.Float64,
			Quest:              quest,
			GenEds:             genEds,
		}
	}
	return typedCourses, nil
}

type timeOfDay struct {
	time.Time
}

func (t *timeOfDay) UnmarshalJSON(data []byte) error {
	var value string

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		return fmt.Errorf("invalid time %q: %w", value, err)
	}

	t.Time = parsed
	return nil
}

func (t timeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format("15:04:05"))
}

type CustomMeetTime struct {
	Days        []MeetDayType `json:"days"`
	TimeStart   *timeOfDay    `json:"time_start"`
	TimeEnd     *timeOfDay    `json:"time_end"`
	PeriodStart *string       `json:"period_start"`
	PeriodEnd   *string       `json:"period_end"`
	Building    *string       `json:"building"`
}

type CustomInstructor struct {
	Name       string   `json:"name"`
	Rating     *float64 `json:"rating"`
	Difficulty *float64 `json:"difficulty"`
	TakeAgain  *float64 `json:"take_again"`
}

type RawListClassesByIDRows []ListClassesByIDRow

type TypedListClassesByIDRow struct {
	ListClassesByIDRow
	MeetTimes   []CustomMeetTime   `json:"meet_times"`
	Instructors []CustomInstructor `json:"instructors"`
}

func (r RawListClassesByIDRows) Typed() ([]TypedListClassesByIDRow, error) {
	typedClasses := make([]TypedListClassesByIDRow, len(r))
	for i, rawClass := range r {
		meetTimesJSON, err := json.Marshal(rawClass.MeetTimes)
		if err != nil {
			return nil, fmt.Errorf("marshal meet times for class %d: %w", rawClass.Number, err)
		}

		var meetTimes []CustomMeetTime
		if err := json.Unmarshal(meetTimesJSON, &meetTimes); err != nil {
			return nil, fmt.Errorf("unmarshal meet times for class %d: %w", rawClass.Number, err)
		}

		instructorsJSON, err := json.Marshal(rawClass.Instructors)
		if err != nil {
			return nil, fmt.Errorf("marshal instructors for class %d: %w", rawClass.Number, err)
		}

		var instructors []CustomInstructor
		if err := json.Unmarshal(instructorsJSON, &instructors); err != nil {
			return nil, fmt.Errorf("unmarshal instructors for class %d: %w", rawClass.Number, err)
		}

		typedClasses[i] = TypedListClassesByIDRow{
			ListClassesByIDRow: rawClass,
			MeetTimes:          meetTimes,
			Instructors:        instructors,
		}
	}
	return typedClasses, nil
}
