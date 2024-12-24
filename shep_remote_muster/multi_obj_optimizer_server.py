from __future__ import print_function
import asyncio
import time
from concurrent import futures

import grpc
import shep_optimizer.shep_optimizer_pb2 as opt_pb
import shep_optimizer.shep_optimizer_pb2_grpc as opt_pb_grpc

from ax.service.ax_client import AxClient
from ax.service.utils.instantiation import ObjectiveProperties
from ax.plot.pareto_utils import compute_posterior_pareto_frontier

lat_target = 500
joules = 0.0
latency = 0.0
itr_vals = [10, 50, 100, 200, 400]
dvfs_vals = [0xc00, 0xe00, 0x1100, 0x1300, 0x1500, 0x1700, 0x1900, 0x1a00]
itr_dict = {'is_ordered': True, 'sort_values': True, 'log_scale': False, 'name': 'itr', 'type': 'choice', 'value_type': 'int', 'values': itr_vals}
dvfs_dict = {'is_ordered': True, 'sort_values': True, 'log_scale': False, 'name': 'dvfs', 'type': 'choice', 'value_type': 'int', 'values': dvfs_vals}
search_space = [itr_dict, dvfs_dict]



def mcd_eval(params):
    itr = params['itr']
    dvfs = params['dvfs']
    with grpc.insecure_channel("localhost:50091") as channel:
        stub = opt_pb_grpc.OptimizeStub(channel)
        new_ctrls = [opt_pb.ControlEntry(knob="dvfs", val=dvfs), opt_pb.ControlEntry(knob="itr-delay", val=itr)]
        response = stub.EvaluateOptimizer(opt_pb.OptimizeRequest(ctrls=new_ctrls))
    for reward in response.rewards:
        if reward.id == "joules":
            joules = reward.val
        if reward.id == "latency":
            latency = reward.val
    print(f'itr = {itr},  dvfs = {dvfs},  joules = {joules}, latency = {latency}\n')
    res = {'mcd_joules': (joules, 0.0), 'mcd_read_99th': (latency, 0.0)}
    return res


async def optimize_func(n_trials, ax_client):
    print(".... Optimize Func ...\n")
    for j in range(n_trials):
        parameters, trial_idx = ax_client.get_next_trial()
        ax_client.complete_trial(trial_index=trial_idx, raw_data=mcd_eval(parameters))

    objectives = ax_client.experiment.optimization_config.objective.objectives
    frontier = compute_posterior_pareto_frontier(
        experiment = ax_client.experiment,
        data = ax_client.experiment.fetch_data(),
        primary_objective = objectives[1].metric,
        secondary_objective = objectives[0].metric,
        absolute_metrics = ["mcd_joules", "mcd_read_99th"],
    )

    print("frontier: ", frontier)

class SetupOptimizeServicer(opt_pb_grpc.SetupOptimizeServicer):
    def __init__(self):
        print(".... .... Optimizer service init..\n")

    async def StartOptimizer(self, request, context) -> opt_pb.StartOptimizerReply:
        print(".... .... StartOptimizer RPC ... request:  ", request, "\n")
        ax_client = AxClient()
        ax_client.create_experiment(
            name="multi_obj_test_mcd",
            parameters=search_space,
            objectives={
                "mcd_joules": ObjectiveProperties(minimize=True),
                "mcd_read_99th": ObjectiveProperties(minimize=True, threshold=500)},
        )
        asyncio.create_task(optimize_func(request.n_trials, ax_client))
        return opt_pb.StartOptimizerReply(done=True) 


async def serve() -> None:
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=1))
    opt_pb_grpc.add_SetupOptimizeServicer_to_server(
        SetupOptimizeServicer(), server
    )
    server.add_insecure_port("[::]:50101")
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    asyncio.run(serve())

