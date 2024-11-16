from __future__ import print_function
import asyncio
import time
from concurrent import futures

import grpc
import shep_optimizer.shep_optimizer_pb2 as opt_pb
import shep_optimizer.shep_optimizer_pb2_grpc as opt_pb_grpc

from ax.service.managed_loop import optimize
from ax.plot.trace import optimization_trace_single_method



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
        print(reward.val)
    joules = reward.val
    print(f'itr = {itr},  dvfs = {dvfs},  joules = {joules}\n')
    #joules, rth = get_joules_latency(itr, dvfs)
    #print(f'itr = {itr},  dvfs = {dvfs},  joules = {joules}, latency = {latency}\n')
    res = {'mcd': (joules, 0.0)}
    return res

async def optimize_func(n_trials):
    print(".... Optimize Func ...\n")
    best_params, values, exp, model = optimize(parameters=search_space, 
                                           evaluation_function = lambda params: mcd_eval(params), 
                                           experiment_name=f'mcd_discrete', 
                                           objective_name='mcd', 
                                           minimize=True, total_trials = n_trials)
    print()
    print(best_params, values)



class SetupOptimizeServicer(opt_pb_grpc.SetupOptimizeServicer):
    def __init__(self):
        print(".... .... Optimizer service init..\n")

    async def StartOptimizer(self, request, context) -> opt_pb.StartOptimizerReply:
        print(".... .... StartOptimizer RPC ... request:  ", request, "\n")
        asyncio.create_task(optimize_func(request.n_trials))
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

