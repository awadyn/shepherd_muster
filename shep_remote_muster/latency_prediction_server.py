from __future__ import print_function
import asyncio
import time
import random
from concurrent import futures

import grpc
import shep_optimizer.shep_optimizer_pb2 as opt_pb
import shep_optimizer.shep_optimizer_pb2_grpc as opt_pb_grpc

itr_vals = [10, 50, 100, 200, 400]
dvfs_vals = [0xc00, 0xe00, 0x1100, 0x1300, 0x1500, 0x1700, 0x1900, 0x1a00]


#def test():


#def train():


async def training_func(n_trials):
    print(".... Latency Prediction Training ...\n")
    for i in range(n_trials):
        print("Training iteration #", i)
        print("Choosing random controls..")
        itr = random.choice(itr_vals)
        dvfs = random.choice(dvfs_vals)
        with grpc.insecure_channel("localhost:50091") as channel:
            stub = opt_pb_grpc.OptimizeStub(channel)
            new_ctrls = [opt_pb.ControlEntry(knob="dvfs", val=dvfs), opt_pb.ControlEntry(knob="itr-delay", val=itr)]
            response = stub.EvaluateOptimizer(opt_pb.OptimizeRequest(ctrls=new_ctrls))
            print("Evaluation response: ", response)
            print()

class SetupOptimizeServicer(opt_pb_grpc.SetupOptimizeServicer):
    def __init__(self):
        print(".... .... Latency Prediction .... service init..\n")

    async def StartOptimizer(self, request, context) -> opt_pb.StartOptimizerReply:
        print(".... .... StartOptimizer RPC ... request:  ", request, "\n")
        asyncio.create_task(training_func(request.n_trials))
        return opt_pb.StartOptimizerReply(done=True)

    async def StopOptimizer(self, request, context) -> opt_pb.StopOptimizerReply:
        print(".... .... StopOptimizer RPC ... request:  ", request, "\n")
        #asyncio.create_task(training_func())
        return opt_pb.StopOptimizerReply(done=True)

       


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

