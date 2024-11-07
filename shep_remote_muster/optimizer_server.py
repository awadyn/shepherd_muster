import asyncio
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
#print(search_space)

async def mcd_run():
    await asyncio.sleep(1)
    print(".... Dummy mcd_run done .... \n")
    return 123

async def mcd_eval(params):
    # first, return new ctrls itr and dvfs to StartOptimizer
    # second, wait for new rewards from OptimizeReward
    # third, feed back new rewards to optimize loop
    itr = params['itr']
    dvfs = params['dvfs']
    joules = await mcd_run()
    #joules, rth = get_joules_latency(itr, dvfs)
    #print(f'itr = {itr},  dvfs = {dvfs},  joules = {joules}, latency = {latency}\n')
    res = {'mcd': (joules, 0.0)}
    return res

best_params, values, exp, model = optimize(parameters=search_space, 
                                           evaluation_function = lambda params: mcd_eval(params), 
                                           experiment_name=f'mcd_discrete', 
                                           objective_name='mcd', 
                                           minimize=True, total_trials = 3)


async def eval_func():
    print(".... Eval Func ...\n")
    return "done"

async def optimize_func():
    print(".... Optimize Func ...\n")
    await eval_func()

async def start_optimize(task):
    await task

class OptimizeServicer(opt_pb_grpc.OptimizeServicer):
    def __init__(self):
        print(".... .... Optimizer service init..\n")

    def StartOptimizer(self, request, context):
        print(".... .... StartOptimizer RPC ... request:  ", request, "\n")
	# async call optimize_func()
#        task = asyncio.create_task(print("HERE"))
#        asyncio.run(start_optimize(task))
        asyncio.run(optimize_func())
	# wait for new ctrls signal
        new_ctrls = [opt_pb.ControlEntry(knob="dvfs", val=0x1500), opt_pb.ControlEntry(knob="itr-delay", val=50)]
        return opt_pb.StartOptimizerReply(done=True, ctrls=new_ctrls)

    def OptimizeReward(self, request, context):
        print(".... .... OptimizeReward RPC ... request:  ", request.rewards, "\n")
	# feed rewards to bayesian optimization loop..
	# send new rewards signal to mcd_eval
        asyncio.run(optimize_func())
        new_ctrls = [opt_pb.ControlEntry(knob="dvfs", val=0x1900), opt_pb.ControlEntry(knob="itr-delay", val=100)]
        return opt_pb.OptimizeRewardReply(done=True, ctrls=new_ctrls)
       

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    opt_pb_grpc.add_OptimizeServicer_to_server(
        OptimizeServicer(), server
    )
    server.add_insecure_port("[::]:50091")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()

