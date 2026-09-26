from ortools.sat.python import cp_model
from redis.asyncio import Redis

from dependencies.redis import ProtoCache
from generated.scheduler.v1 import (
    GenerateSchedulesRequest,
    GenerateSchedulesResponse,
    Schedule,
    SchedulerServiceBase,
)

from .model import build_model
from .objectives import add_objective
from .schedule import extract_schedule

MAX_SOLVE_TIME_SECONDS = 60


class SchedulerService(SchedulerServiceBase):
    def __init__(self, redis: Redis):
        self.schedule_cache = ProtoCache(
            redis,
            "schedules:v1",
            GenerateSchedulesResponse,
        )

    async def generate_schedules(
        self,
        message: GenerateSchedulesRequest,
    ) -> GenerateSchedulesResponse:
        response = await self.schedule_cache.get(message)
        if response is not None:
            return response

        if message.sort_by is None:
            schedules = self.generate_feasible_schedules(
                message,
            )
        else:
            schedules = self.generate_optimized_schedules(
                message,
            )

        response = GenerateSchedulesResponse(
            schedules=schedules,
        )
        await self.schedule_cache.set(message, response)
        return response

    def generate_feasible_schedules(
        self,
        message: GenerateSchedulesRequest,
    ) -> list[Schedule]:
        model, variables = build_model(message)

        solver = cp_model.CpSolver()
        solver.parameters.max_time_in_seconds = MAX_SOLVE_TIME_SECONDS

        class Callback(cp_model.CpSolverSolutionCallback):
            def __init__(self):
                super().__init__()
                self.schedules: list[Schedule] = []

            def on_solution_callback(self):
                self.schedules.append(
                    extract_schedule(
                        self,
                        message,
                        variables,
                    )
                )

        callback = Callback()

        solver.parameters.enumerate_all_solutions = True

        solver.solve(
            model,
            callback,
        )

        return callback.schedules

    def generate_optimized_schedules(
        self,
        message: GenerateSchedulesRequest,
    ) -> list[Schedule]:
        schedules = []
        previous_value = None
        remaining_time = MAX_SOLVE_TIME_SECONDS

        while remaining_time > 0:
            model, variables = build_model(message)

            objective, minimize = add_objective(message, model, variables)

            if previous_value is not None:
                if minimize:
                    model.add(objective > previous_value)
                else:
                    model.add(objective < previous_value)

            if minimize:
                model.minimize(objective)
            else:
                model.maximize(objective)

            solver = cp_model.CpSolver()
            solver.parameters.max_time_in_seconds = remaining_time

            status = solver.solve(model)
            remaining_time -= solver.wall_time

            if remaining_time <= 0:
                break
            if status not in (cp_model.OPTIMAL, cp_model.FEASIBLE):
                break

            current_value = solver.value(objective)

            model, variables = build_model(message)

            objective, _ = add_objective(
                message,
                model,
                variables,
            )

            model.add(objective == current_value)

            solver = cp_model.CpSolver()
            solver.parameters.max_time_in_seconds = remaining_time

            class Callback(cp_model.CpSolverSolutionCallback):
                def __init__(self, variables: dict[tuple[str, str], cp_model.IntVar]):
                    super().__init__()
                    self.variables = variables
                    self.schedules: list[Schedule] = []

                def on_solution_callback(self):
                    self.schedules.append(
                        extract_schedule(
                            self,
                            message,
                            self.variables,
                        )
                    )

            callback = Callback(variables)

            solver.parameters.enumerate_all_solutions = True

            solver.solve(
                model,
                callback,
            )

            remaining_time -= solver.wall_time

            schedules.extend(callback.schedules)

            previous_value = current_value

        return schedules
