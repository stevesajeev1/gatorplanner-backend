package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/rangerelation"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/textquerytype"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
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
	value := types.Float64(*rule.NumberValue)

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
	CourseCode string                       `json:"course_code"`
	Classes    []sqlc.TypedSearchClassesRow `json:"classes"`
}

func (r *ClassesESRepository) Search(
	ctx context.Context,
	termID int64,
	search *string,
	filter *search.Filter,
) ([]SearchClassResult, error) {
	query := esdsl.NewBoolQuery()

	if search != nil {
		query = query.Must(buildSearchQuery(*search))
	}

	termFilter := esdsl.NewTermQuery(
		"term_id",
		esdsl.NewFieldValue().Int64(termID),
	)

	if filter != nil {
		query = query.Filter(termFilter, buildFilterQuery(*filter))
	} else {
		query = query.Filter(termFilter)
	}

	data, _ := query.QueryCaster().MarshalJSON()
	fmt.Println(string(data))

	results, err := r.es.Search().
		Index("classes").
		Query(query).
		Collapse(
			esdsl.NewFieldCollapse().
				Field("course_code.keyword").
				InnerHits(esdsl.NewInnerHits().Name("classes").Size(100)),
		).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch search: %w", err)
	}

	out := make([]SearchClassResult, len(results.Hits.Hits))
	for i, hit := range results.Hits.Hits {
		out[i] = SearchClassResult{}

		var course struct {
			CourseCode string `json:"course_code"`
		}
		if err := json.Unmarshal(hit.Source_, &course); err != nil {
			return nil, fmt.Errorf("unmarshal course: %w", err)
		}
		out[i].CourseCode = course.CourseCode

		innerHits, ok := hit.InnerHits["classes"]
		if !ok {
			continue
		}

		out[i].Classes = make([]sqlc.TypedSearchClassesRow, len(innerHits.Hits.Hits))

		for j, classHit := range innerHits.Hits.Hits {
			if err := json.Unmarshal(classHit.Source_, &out[i].Classes[j]); err != nil {
				return nil, fmt.Errorf("unmarshal class: %w", err)
			}
		}
	}
	return out, nil
}
