package main

/**************************************/


func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.10.1.2", ip_idx: -1, ncores: 4, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081}}
//	nodes := []node{{ip: "localhost", ip_idx: -1, ncores: 16, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081}}
//	nodes := []node{{ip: "localhost", ip_idx: 0, ncores: 1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081},
//			{ip: "localhost", ip_idx: 1, ncores: 1, pulse_port: 50052, log_port:50062, ctrl_port: 50072, coordinate_port: 50082}}

	bayopt_main(nodes)
//	flink_main(nodes)
}





