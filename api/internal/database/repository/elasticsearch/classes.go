package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/rangerelation"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/textquerytype"
	"github.com/google/uuid"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/classes/search"
)

type ClassesESRepository struct {
	es *dependencies.ES
}

func NewClassesESRepository(es *dependencies.ES) *ClassesESRepository {
	return &ClassesESRepository{es: es}
}

func buildSearchQuery(search string) types.QueryVariant {
	return esdsl.NewBoolQuery().
		Should(
			esdsl.NewMultiMatchQuery(search).
				Fields(
					"course_name^8",
					"course_code^7",
					"course_description^5",
					"course_department^4",
					"course_gen_eds^2",
					"course_quest^2",
					"note^1.5",
					"course_prerequisites",
				).
				Type(textquerytype.Bestfields).
				Fuzziness(esdsl.NewFuzziness().String("AUTO")).
				MinimumShouldMatch(esdsl.NewMinimumShouldMatch().String("75%")),

			esdsl.NewNestedQuery(
				esdsl.NewMatchQuery("instructors.name", search).
					Fuzziness(esdsl.NewFuzziness().String("AUTO")).
					Boost(3),
			).Path("instructors"),

			esdsl.NewNestedQuery(
				esdsl.NewMatchQuery("meet_times.building", search).
					Fuzziness(esdsl.NewFuzziness().String("AUTO")).
					Boost(2),
			).Path("meet_times"),

			esdsl.NewTermQuery(
				"course_code.keyword",
				esdsl.NewFieldValue().String(search),
			).Boost(20),

			esdsl.NewTermQuery(
				"course_department.keyword",
				esdsl.NewFieldValue().String(search),
			).Boost(10),
		).
		MinimumShouldMatch(esdsl.NewMinimumShouldMatch().Int(1))
}

func buildFilterQuery(filter search.Filter) types.QueryVariant {
	query := esdsl.NewBoolQuery()

	ruleQueries := make([]types.QueryVariant, len(filter.Rules))
	for i, node := range filter.Rules {
		switch node := node.(type) {
		case search.Filter:
			ruleQueries[i] = buildFilterQuery(node)

		case search.Rule:
			ruleQueries[i] = buildRuleQuery(node)
		}
	}

	switch filter.Glue {
	case search.GlueAnd:
		query = query.Must(ruleQueries...)

	case search.GlueOr:
		query = query.Should(ruleQueries...).
			MinimumShouldMatch(
				esdsl.NewMinimumShouldMatch().Int(1),
			)
	}

	return query
}

func buildRuleQuery(rule search.Rule) types.QueryVariant {
	var query types.QueryVariant

	switch rule.Type {
	case search.FieldTypeText:
		query = buildTextRuleQuery(rule)

	case search.FieldTypeNumber:
		query = buildNumberRuleQuery(rule)

	case search.FieldTypeTime:
		query = buildNumberRuleQuery(rule)

	case search.FieldTypeBoolean:
		query = buildBooleanRuleQuery(rule)
	}

	return wrapNestedQuery(rule.Field, query)
}

func buildTextRuleQuery(rule search.Rule) types.QueryVariant {
	field := keywordField(rule.Field)

	switch rule.Filter {
	case search.FieldFilterEqual:
		return esdsl.NewTermQuery(
			field,
			esdsl.NewFieldValue().String(*rule.TextValue),
		)

	case search.FieldFilterNotEqual:
		return esdsl.NewBoolQuery().
			MustNot(
				esdsl.NewTermQuery(
					field,
					esdsl.NewFieldValue().String(*rule.TextValue),
				),
			)
	}

	return nil
}

func keywordField(field search.Field) string {
	switch field {
	case search.FieldCourseCode,
		search.FieldCourseName,
		search.FieldCourseDepartment,
		search.FieldCourseGenEds,
		search.FieldCourseQuest,
		search.FieldCourseMeetBuilding,
		search.FieldInstructorName:

		return string(field) + ".keyword"

	default:
		return string(field)
	}
}

func buildNumberRuleQuery(rule search.Rule) types.QueryVariant {
	switch rule.Field {
	case search.FieldCourseCredits:
		return buildRangeNumberRuleQuery(rule)
	default:
		return buildScalarNumberRuleQuery(rule)
	}
}

func buildRangeNumberRuleQuery(rule search.Rule) types.QueryVariant {
	value := types.Float64(*rule.NumberValue)

	query := esdsl.NewNumberRangeQuery(string(rule.Field))

	switch rule.Filter {
	case search.FieldFilterEqual:
		return query.
			Gte(value).
			Lte(value).
			Relation(rangerelation.Contains)

	case search.FieldFilterNotEqual:
		return esdsl.NewBoolQuery().
			MustNot(
				esdsl.NewNumberRangeQuery(string(rule.Field)).
					Gte(value).
					Lte(value).
					Relation(rangerelation.Contains),
			)

	case search.FieldFilterGreater:
		return query.Gt(value)

	case search.FieldFilterLess:
		return query.Lt(value)

	case search.FieldFilterGreaterOrEqual:
		return query.Gte(value)

	case search.FieldFilterLessOrEqual:
		return query.Lte(value)
	}

	return nil
}

func buildScalarNumberRuleQuery(rule search.Rule) types.QueryVariant {
	var value types.Float64
	switch rule.Type {
	case search.FieldTypeNumber:
		switch rule.Field {
		case search.FieldCourseMeetPeriodStart, search.FieldCourseMeetPeriodEnd:
			period := *rule.TextValue

			if after, ok := strings.CutPrefix(period, "E"); ok {
				n, _ := strconv.Atoi(after)
				value = types.Float64(11 + n)
			} else {
				n, _ := strconv.Atoi(period)
				value = types.Float64(n)
			}
		default:
			value = types.Float64(*rule.NumberValue)
		}
	case search.FieldTypeTime:
		time := *rule.TimeValue
		value = types.Float64(time.Hour*60 + time.Minute)
	}

	switch rule.Filter {
	case search.FieldFilterEqual:
		return esdsl.NewTermQuery(
			string(rule.Field),
			esdsl.NewFieldValue().Float64(value),
		)

	case search.FieldFilterNotEqual:
		return esdsl.NewBoolQuery().
			MustNot(
				esdsl.NewTermQuery(
					string(rule.Field),
					esdsl.NewFieldValue().Float64(value),
				),
			)

	case search.FieldFilterGreater:
		return esdsl.NewNumberRangeQuery(string(rule.Field)).
			Gt(value)

	case search.FieldFilterLess:
		return esdsl.NewNumberRangeQuery(string(rule.Field)).
			Lt(value)

	case search.FieldFilterGreaterOrEqual:
		return esdsl.NewNumberRangeQuery(string(rule.Field)).
			Gte(value)

	case search.FieldFilterLessOrEqual:
		return esdsl.NewNumberRangeQuery(string(rule.Field)).
			Lte(value)
	}

	return nil
}

func buildBooleanRuleQuery(rule search.Rule) types.QueryVariant {
	switch rule.Filter {
	case search.FieldFilterEqual:
		return esdsl.NewTermQuery(
			string(rule.Field),
			esdsl.NewFieldValue().Bool(*rule.BooleanValue),
		)

	case search.FieldFilterNotEqual:
		return esdsl.NewBoolQuery().
			MustNot(
				esdsl.NewTermQuery(
					string(rule.Field),
					esdsl.NewFieldValue().Bool(*rule.BooleanValue),
				),
			)
	}

	return nil
}

func wrapNestedQuery(field search.Field, query types.QueryVariant) types.QueryVariant {
	switch field {
	case search.FieldCourseMeetDays,
		search.FieldCourseMeetTimeStart,
		search.FieldCourseMeetTimeEnd,
		search.FieldCourseMeetPeriodStart,
		search.FieldCourseMeetPeriodEnd,
		search.FieldCourseMeetBuilding:

		return esdsl.NewNestedQuery(query).
			Path("meet_times")

	case search.FieldInstructorName,
		search.FieldInstructorRating,
		search.FieldInstructorDifficulty,
		search.FieldInstructorTakeAgain:

		return esdsl.NewNestedQuery(query).
			Path("instructors")

	default:
		return query
	}
}

type SearchClassResult struct {
	Total uint
	Items []*SearchClassResultItem
}

type SearchClassResultItem struct {
	CourseID uuid.UUID
	ClassIDs []uuid.UUID
}

func (r *ClassesESRepository) Search(
	ctx context.Context,
	termID int32,
	search *string,
	filter *search.Filter,
	limit uint,
	offset uint,
) (*SearchClassResult, error) {
	query := esdsl.NewBoolQuery()

	if search != nil {
		query = query.Must(buildSearchQuery(*search))
	}

	termFilter := esdsl.NewTermQuery(
		"term_id",
		esdsl.NewFieldValue().Int64(int64(termID)),
	)

	if filter != nil {
		query = query.Filter(termFilter, buildFilterQuery(*filter))
	} else {
		query = query.Filter(termFilter)
	}

	res, err := r.es.Search().
		Index("classes").
		Query(query).
		Collapse(
			esdsl.NewFieldCollapse().
				Field("course_id").
				InnerHits(
					esdsl.NewInnerHits().
						Name("classes").
						Size(100).
						Source_(esdsl.NewSourceFilter().Excludes("*")),
				),
		).
		From(int(offset)).
		Size(int(limit)).
		AddAggregation(
			"total_courses",
			esdsl.NewCardinalityAggregation().
				Field("course_id"),
		).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search: %w", err)
	}

	total := uint(res.Aggregations["total_courses"].(*types.CardinalityAggregate).Value)

	items := make([]*SearchClassResultItem, len(res.Hits.Hits))
	for i, courseHit := range res.Hits.Hits {
		var course struct {
			CourseID uuid.UUID `json:"course_id"`
		}
		if err := json.Unmarshal(courseHit.Source_, &course); err != nil {
			return nil, fmt.Errorf("unmarshal course: %w", err)
		}

		innerHits, ok := courseHit.InnerHits["classes"]
		if !ok {
			continue
		}

		classIDs := make([]uuid.UUID, len(innerHits.Hits.Hits))
		for j, classHit := range innerHits.Hits.Hits {
			id, err := uuid.Parse(*classHit.Id_)
			if err != nil {
				return nil, fmt.Errorf("parse class ID: %w", err)
			}

			classIDs[j] = id
		}

		items[i] = &SearchClassResultItem{
			CourseID: course.CourseID,
			ClassIDs: classIDs,
		}
	}

	return &SearchClassResult{
		Total: total,
		Items: items,
	}, nil
}
