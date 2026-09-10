from ortools.sat.python import cp_model

from generated.scheduler.v1 import (
    GenerateSchedulesRequest,
    GenerateSchedulesResponse,
    Schedule,
    SelectedClass,
    SchedulerServiceBase,
)


def meetings_overlap(a, b) -> bool:
    """
    Returns True if two meetings overlap on at least one common day.
    """

    common_days = set(a.days) & set(b.days)

    if not common_days:
        return False

    return (
        a.start_minute < b.end_minute
        and b.start_minute < a.end_minute
    )


class SchedulerService(SchedulerServiceBase):
    async def generate_schedules(
        self,
        message: GenerateSchedulesRequest,
    ) -> GenerateSchedulesResponse:
        model = cp_model.CpModel()

        # variable[(course_id, class_id)] = whether this class is selected.
        variables = {}

        # Create one Boolean variable for every possible class.
        for course in message.courses:
            for class_ in course.classes:
                variables[(course.course_id, class_.class_id)] = (
                    model.new_bool_var(
                        f"course_{course.course_id}_class_{class_.class_id}"
                    )
                )

        # Exactly one class must be selected for each course.
        for course in message.courses:
            class_variables = [
                variables[(course.course_id, class_.class_id)]
                for class_ in course.classes
            ]

            model.add_exactly_one(class_variables)

        # Prevent conflicting classes from being selected together.
        all_classes = [
            (course, class_)
            for course in message.courses
            for class_ in course.classes
        ]

        for i, (course_a, class_a) in enumerate(all_classes):
            for course_b, class_b in all_classes[i + 1:]:
                # A course cannot conflict with itself because exactly
                # one class is selected, so skip it.
                if course_a.course_id == course_b.course_id:
                    continue

                # Check every meeting pair.
                conflict = any(
                    meetings_overlap(meeting_a, meeting_b)
                    for meeting_a in class_a.meetings
                    for meeting_b in class_b.meetings
                )

                if conflict:
                    model.add(
                        variables[(course_a.course_id, class_a.class_id)]
                        + variables[(course_b.course_id, class_b.class_id)]
                        <= 1
                    )

        solver = cp_model.CpSolver()

        # Find all feasible solutions.
        schedules = []

        class ScheduleCallback(cp_model.CpSolverSolutionCallback):
            def __init__(self, limit: int):
                super().__init__()
                self.limit = limit
                self.schedules = []

            def on_solution_callback(self):
                if len(self.schedules) >= self.limit:
                    self.stop_search()
                    return

                selected_classes = []

                for course in message.courses:
                    for class_ in course.classes:
                        variable = variables[
                            (course.course_id, class_.class_id)
                        ]

                        if self.value(variable):
                            selected_classes.append(
                                SelectedClass(
                                    course_id=course.course_id,
                                    class_id=class_.class_id,
                                )
                            )

                self.schedules.append(
                    Schedule(classes=selected_classes)
                )

        callback = ScheduleCallback(limit=100)

        solver.parameters.enumerate_all_solutions = True
        solver.solve(model, callback)

        schedules = callback.schedules

        return GenerateSchedulesResponse(
            schedules=schedules,
        )