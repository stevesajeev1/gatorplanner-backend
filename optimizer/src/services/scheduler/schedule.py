from ortools.sat.python import cp_model

from generated.scheduler.v1 import (
    GenerateSchedulesRequest,
    Schedule,
    SelectedClass,
)


def extract_schedule(
    solver: cp_model.CpSolverSolutionCallback,
    message: GenerateSchedulesRequest,
    variables: dict[tuple[str, str], cp_model.IntVar],
) -> Schedule:
    selected_classes = []

    for course in message.courses:
        for cls in course.classes:
            variable = variables[(course.course_id, cls.class_id)]

            if solver.value(variable):
                selected_classes.append(
                    SelectedClass(
                        course_id=course.course_id,
                        class_id=cls.class_id,
                    )
                )

    return Schedule(
        classes=selected_classes,
    )
