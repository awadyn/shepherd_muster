from concurrent import futures
import grpc
import shep_optimizer.shep_optimizer_pb2 as opt_pb
import shep_optimizer.shep_optimizer_pb2_grpc as opt_pb_grpc


class OptimizeServicer(opt_pb_grpc.OptimizeServicer):
    def __init__(self):
        print(".... .... Optimizer service init..\n")

    def StartOptimizer(self, request, context):
        print(".... .... StartOptimizer RPC ... request:  ", request, "\n")
        new_ctrls = [opt_pb.ControlEntry(knob="dvfs", val=0x1500), opt_pb.ControlEntry(knob="itr-delay", val=50)]
        return opt_pb.StartOptimizerReply(done=True, ctrls=new_ctrls)

    def OptimizeReward(self, request, context):
        print(".... .... OptimizeReward RPC ... request:  ", request.rewards, "\n")
	# feed rewards to bayesian optimization loop..
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

