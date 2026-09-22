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
	GenEds     []GenEd `json:"gen_eds" nullable:"false"`
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

		genEds := []GenEd{}
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
	Days        []MeetDayType `json:"days" nullable:"false"`
	TimeStart   timeOfDay     `json:"time_start"`
	TimeEnd     timeOfDay     `json:"time_end"`
	PeriodStart string        `json:"period_start"`
	PeriodEnd   string        `json:"period_end"`
	Building    string        `json:"building"`
}

type CustomInstructor struct {
	Name       string   `json:"name"`
	RMPId      *int32   `json:"rmp_id"`
	Rating     *float64 `json:"rating"`
	Difficulty *float64 `json:"difficulty"`
	TakeAgain  *float64 `json:"take_again"`
}

type RawListClassesByIDRows []ListClassesByIDRow

type TypedListClassesByIDRow struct {
	ListClassesByIDRow
	MeetTimes   []CustomMeetTime   `json:"meet_times" nullable:"false"`
	Instructors []CustomInstructor `json:"instructors" nullable:"false"`
}

func (r RawListClassesByIDRows) Typed() ([]TypedListClassesByIDRow, error) {
	typedClasses := make([]TypedListClassesByIDRow, len(r))
	for i, rawClass := range r {
		meetTimesJSON, err := json.Marshal(rawClass.MeetTimes)
		if err != nil {
			return nil, fmt.Errorf("marshal meet times for class %d: %w", rawClass.Number, err)
		}

		meetTimes := []CustomMeetTime{}
		if err := json.Unmarshal(meetTimesJSON, &meetTimes); err != nil {
			return nil, fmt.Errorf("unmarshal meet times for class %d: %w", rawClass.Number, err)
		}

		instructorsJSON, err := json.Marshal(rawClass.Instructors)
		if err != nil {
			return nil, fmt.Errorf("marshal instructors for class %d: %w", rawClass.Number, err)
		}

		instructors := []CustomInstructor{}
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

type RawListBuildingsRows []ListBuildingsRow

type TypedListBuildingsRow struct {
	ListBuildingsRow
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (r RawListBuildingsRows) Typed() ([]TypedListBuildingsRow, error) {
	typedBuildings := make([]TypedListBuildingsRow, len(r))
	for i, rawBuilding := range r {
		var latitude *float64 = nil
		if rawBuilding.Latitude.Valid {
			value, err := rawBuilding.Latitude.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("convert latitude for building %s: %w", rawBuilding.Name, err)
			}
			latitude = &value.Float64
		}
		var longitude *float64 = nil
		if rawBuilding.Longitude.Valid {
			value, err := rawBuilding.Longitude.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("convert longitude for building %s: %w", rawBuilding.Name, err)
			}
			longitude = &value.Float64
		}

		typedBuildings[i] = TypedListBuildingsRow{
			ListBuildingsRow: rawBuilding,
			Latitude:         latitude,
			Longitude:        longitude,
		}
	}
	return typedBuildings, nil
}
