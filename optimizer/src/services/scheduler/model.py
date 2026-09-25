from itertools import combinations

from ortools.sat.python import cp_model

from generated.scheduler.v1 import (
    Class,
    Day,
    GenerateSchedulesRequest,
    Meeting,
)


def meetings_overlap(a: Meeting, b: Meeting) -> bool:
    common_days = set(a.days) & set(b.days)
    if not common_days:
        return False

    return (
        a.start_minute < b.end_minute
        and b.start_minute < a.end_minute
    )


def class_allowed_on_days(
    cls: Class,
    restricted_days: set[Day],
) -> bool:
    if not restricted_days:
        return True

    for meeting in cls.meetings:
        if any(day in restricted_days for day in meeting.days):
            return False
    return True


def build_model(
    message: GenerateSchedulesRequest,
) -> tuple[cp_model.CpModel, dict[tuple[str, str], cp_model.IntVar]]:
    model = cp_model.CpModel()

    # variable[(course_id, class_id)] = class is selected
    variables: dict[tuple[str, str], cp_model.IntVar] = {}

    for course in message.courses:
        for cls in course.classes:
            variables[(course.course_id, cls.class_id)] = (
                model.new_bool_var(
                    f"course_{course.course_id}_class_{cls.class_id}"
                )
            )

    # one class selected for each course
    for course in message.courses:
        class_variables = [
            variables[(course.course_id, cls.class_id)]
            for cls in course.classes
        ]

        model.add_exactly_one(class_variables)

    restricted_days = set(message.day_restrictions)

    if restricted_days:
        for course in message.courses:
            for cls in course.classes:
                if not class_allowed_on_days(
                    cls,
                    restricted_days,
                ):
                    model.add(
                        variables[(course.course_id, cls.class_id)] == 0
                    )

    # prevent conflicting classes from being selected together
    for course_a, course_b in combinations(message.courses, 2):
        for class_a in course_a.classes:
            for class_b in course_b.classes:
                if any(
                    meetings_overlap(ma, mb)
                    for ma in class_a.meetings
                    for mb in class_b.meetings
                ):
                    model.add_at_most_one(
                        variables[(course_a.course_id, class_a.class_id)],
                        variables[(course_b.course_id, class_b.class_id)]
                    )

    return model, variables