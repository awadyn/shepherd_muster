from __future__ import print_function
import asyncio
from concurrent import futures

import time
import sys
import os
import pandas as pd

import grpc
import shep_log_processor.shep_log_processor_pb2 as opt_pb
import shep_log_processor.shep_log_processor_pb2_grpc as opt_pb_grpc

from google.protobuf.any_pb2 import Any
from google.protobuf.wrappers_pb2 import Int64Value
from google.protobuf.wrappers_pb2 import StringValue

port_server = sys.argv[1]
port_client = sys.argv[2]

logs_dir = ""
logs_idx = {}
cols = ["rx_bytes", "instructions", "cycles", "ref_cycles", "llc_miss", "timestamp", "core"]
merged_log = pd.DataFrame(columns=cols)

async def log_proc_func(request):
    print(".... Log Processing Function .... ", request, "\n")

def initializer(dir_path):
    global logs_dir
    global logs_idx
    logs_dir = dir_path
    for file in os.listdir(logs_dir):
        core = file.split('-')[3]                   # log file name: ixgbe-log-core-<X>-<REMOTE_IP>
        logs_idx[core] = 0
    print(".... INITIALIZED: logs_dir = ", logs_dir, " -- file seekers: ", logs_idx)

async def merge_log(log_file, core):
    global cols
    global merged_log
    global logs_idx
    df = pd.read_csv(log_file, sep = ' ', names = cols, skiprows = logs_idx[core])
    df["core"] = [core] * df.shape[0]
    concat_log = pd.concat([merged_log, df])
    merged_log = concat_log.sort_values(by="timestamp", ascending=True)
    merged_log = merged_log.reset_index(drop=True)
    logs_idx[core] += df.shape[0]

async def merged_log_eval(n):
    global merged_log
    elapsed_time = 0
    loop = asyncio.get_running_loop()
    start_time = loop.time()
    i = 0
    while i < n:
        if elapsed_time > 3:
            if merged_log.shape[0] > 0:
                rx_bytes_sum = sum(merged_log["rx_bytes"].values)
                rx_bytes_median = sorted(merged_log["rx_bytes"].values)[int(merged_log.shape[0]/2)]
                arg0 = Any()
                arg0.Pack(Int64Value(value=rx_bytes_sum))
                arg1 = Any()
                arg1.Pack(Int64Value(value=rx_bytes_median))
                args = [arg0, arg1]
                print("READY TO RPC -- ", rx_bytes_sum, rx_bytes_median)
                with grpc.insecure_channel("localhost:50091") as channel:
                    stub = opt_pb_grpc.LogStatsMessengerStub(channel)
                    response = stub.EvaluateLogStats(opt_pb.EvaluateLogStatsRequest(args=args))
                    print("DONE RPC -- ", response)
                merged_log = pd.DataFrame(columns=cols)
                start_time = loop.time()
        end_time = loop.time()
        elapsed_time = end_time - start_time
        i += 1
        await asyncio.sleep(0.5)


class LogProcessorServicer(opt_pb_grpc.LogProcessorServicer):
    def __init__(self):
        print(".... .... Setup Optimizer INIT..\n")

    async def StartOptimizer(self, request, context) -> opt_pb.StartOptimizerReply:
        path = StringValue()
        request.args[0].Unpack(path)
        initializer(path.value)
        asyncio.create_task(merged_log_eval(10000))
        print(".... .... Log Processing Optimizer .... StartOptimizer RPC .... logs_dir:  ", logs_dir, "\n")
        return opt_pb.StartOptimizerReply(done=True)

    async def ProcessLog(self, request, context) -> opt_pb.ProcessLogReply:   
        log_id = StringValue()
        request.args[0].Unpack(log_id)
        filename = logs_dir + log_id.value
        core = log_id.value.split('-')[3]
        asyncio.create_task(merge_log(filename, core))
        return opt_pb.ProcessLogReply(done=True)


async def serve() -> None:
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=1))
    opt_pb_grpc.add_LogProcessorServicer_to_server(
        LogProcessorServicer(), server
    )
    server.add_insecure_port("[::]:" + port_server)
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    asyncio.run(serve())

