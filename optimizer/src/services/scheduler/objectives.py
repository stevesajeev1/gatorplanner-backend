from typing import Callable

from ortools.sat.python import cp_model

from generated.scheduler.v1 import (
    Day,
    GenerateSchedulesRequest,
    SortBy,
    Class,
)

MINUTES_PER_DAY = 24 * 60
MAX_MINUTE = MINUTES_PER_DAY


def class_earliest_start(cls: Class) -> int:
    return min(
        meeting.start_minute
        for meeting in cls.meetings
    )


def class_latest_start(cls: Class) -> int:
    return max(
        meeting.start_minute
        for meeting in cls.meetings
    )


def class_earliest_end(cls: Class) -> int:
    return min(
        meeting.end_minute
        for meeting in cls.meetings
    )


def class_latest_end(cls: Class) -> int:
    return max(
        meeting.end_minute
        for meeting in cls.meetings
    )


def add_selected_value(
    model: cp_model.CpModel,
    selected: cp_model.IntVar,
    value: int,
    name: str,
    default: int,
) -> cp_model.IntVar:
    result = model.new_int_var(
        0,
        MAX_MINUTE,
        name,
    )

    model.add(
        result == value
    ).only_enforce_if(selected)

    model.add(
        result == default
    ).only_enforce_if(selected.Not())

    return result


def add_min_class_value_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
    value_fn: Callable[[Class], int],
    name: str,
) -> cp_model.IntVar:
    values: list[cp_model.IntVar] = []

    for course in message.courses:
        for cls in course.classes:
            selected = variables[
                (course.course_id, cls.class_id)
            ]

            if not cls.meetings:
                value = MAX_MINUTE
            else:
                value = value_fn(cls)

            values.append(
                add_selected_value(
                    model,
                    selected,
                    value,
                    f"{name}_{course.course_id}_{cls.class_id}",
                    MAX_MINUTE,
                )
            )

    objective = model.new_int_var(
        0,
        MAX_MINUTE,
        name,
    )

    model.add_min_equality(
        objective,
        values,
    )

    return objective


def add_max_class_value_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
    value_fn: Callable[[Class], int],
    name: str,
) -> cp_model.IntVar:
    values: list[cp_model.IntVar] = []

    for course in message.courses:
        for cls in course.classes:
            selected = variables[
                (course.course_id, cls.class_id)
            ]

            if not cls.meetings:
                value = 0
            else:
                value = value_fn(cls)

            values.append(
                add_selected_value(
                    model,
                    selected,
                    value,
                    f"{name}_{course.course_id}_{cls.class_id}",
                    0,
                )
            )

    objective = model.new_int_var(
        0,
        MAX_MINUTE,
        name,
    )

    model.add_max_equality(
        objective,
        values,
    )

    return objective


def add_earliest_start_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    return add_min_class_value_objective(
        message,
        model,
        variables,
        class_earliest_start,
        "schedule_earliest_start",
    )


def add_latest_start_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    return add_max_class_value_objective(
        message,
        model,
        variables,
        class_latest_start,
        "schedule_latest_start",
    )


def add_earliest_end_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    return add_min_class_value_objective(
        message,
        model,
        variables,
        class_earliest_end,
        "schedule_earliest_end",
    )


def add_latest_end_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    return add_max_class_value_objective(
        message,
        model,
        variables,
        class_latest_end,
        "schedule_latest_end",
    )


def add_most_compact_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    earliest_start = add_earliest_start_objective(
        message,
        model,
        variables,
    )

    latest_end = add_latest_end_objective(
        message,
        model,
        variables,
    )

    compact = model.new_int_var(
        0,
        MAX_MINUTE,
        "schedule_compactness",
    )

    model.add(
        compact == latest_end - earliest_start
    )

    return compact


def add_fewest_days_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    day_used = {
        day: model.new_bool_var(f"day_used_{day.name}")
        for day in Day
    }

    for course in message.courses:
        for cls in course.classes:
            selected = variables[
                (course.course_id, cls.class_id)
            ]

            for day in day_used:
                if any(day in meeting.days for meeting in cls.meetings):
                    model.add(day_used[day] >= selected)

    objective = model.new_int_var(
        0,
        len(day_used),
        "schedule_day_count",
    )

    model.add(
        objective == sum(day_used.values())
    )

    return objective


def add_most_balanced_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    classes_per_day = {
        day: model.new_int_var(0, len(message.courses), f"classes_on_day_{day.name}")
        for day in Day
    }

    for day in Day:
        class_indicators = []

        for course in message.courses:
            for cls in course.classes:
                selected = variables[
                    (course.course_id, cls.class_id)
                ]

                if any(
                    day in meeting.days
                    for meeting in cls.meetings
                ):
                    class_indicators.append(selected)

        model.add(
            classes_per_day[day] == sum(class_indicators)
            if class_indicators
            else classes_per_day[day] == 0
        )

    objective = model.new_int_var(
        0,
        len(message.courses),
        "maximum_classes_per_day",
    )

    model.add_max_equality(
        objective,
        list(classes_per_day.values()),
    )

    return objective


def add_instructor_rating_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> cp_model.IntVar:
    rating_terms = []

    for course in message.courses:
        for cls in course.classes:
            selected = variables[
                (course.course_id, cls.class_id)
            ]

            # assume a rating of 3.0 for classes without a rating
            if cls.avg_instructor_rating is None:
                rating = 3.0 * 100
            else:
                rating = round(cls.avg_instructor_rating * 100)

            rating_terms.append(rating * selected)

    objective = model.new_int_var(
        0,
        500 * len(message.courses),
        "instructor_rating",
    )

    model.add(objective == sum(rating_terms))

    return objective


def add_objective(
    message: GenerateSchedulesRequest,
    model: cp_model.CpModel,
    variables: dict[tuple[str, str], cp_model.IntVar]
) -> tuple[cp_model.IntVar, bool]:
    sort_by = message.sort_by

    match sort_by:
        case SortBy.EARLIEST_START:
            return (
                add_earliest_start_objective(
                    message,
                    model,
                    variables,
                ),
                True,
            )
        case SortBy.LATEST_START:
            return (
                add_latest_start_objective(
                    message,
                    model,
                    variables,
                ),
                False,
            )
        case SortBy.EARLIEST_END:
            return (
                add_earliest_end_objective(
                    message,
                    model,
                    variables,
                ),
                True,
            )
        case SortBy.LATEST_END:
            return (
                add_latest_end_objective(
                    message,
                    model,
                    variables,
                ),
                False,
            )
        case SortBy.MOST_COMPACT:
            return (
                add_most_compact_objective(
                    message,
                    model,
                    variables,
                ),
                True,
            )
        case SortBy.FEWEST_DAYS:
            return (
                add_fewest_days_objective(
                    message,
                    model,
                    variables,
                ),
                True,
            )
        case SortBy.MOST_BALANCED:
            return (
                add_most_balanced_objective(
                    message,
                    model,
                    variables,
                ),
                True,
            )
        case SortBy.INSTRUCTOR_RATING:
            return (
                add_instructor_rating_objective(
                    message,
                    model,
                    variables,
                ),
                False,
            )
        case _:
            raise ValueError("SortBy")