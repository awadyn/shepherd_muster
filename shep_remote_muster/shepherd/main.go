package main

import (
	//"fmt"
)

/**************************************/

var specialize_on = true
var optimize_on bool = true
//var optimize_on bool = false
//var debug bool = true
var debug bool = false

var rx_bytes_median uint64 = 0

func main() {
	target_resources := setup_target_resources_cX220(16)

//	optimizer_server_ports := []int{50091}
//	optimizer_client_ports := []int{50101}

	nodes := []node{{ip: "10.10.1.2", ip_idx: -1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081, resources: target_resources}}
//	nodes := []node{{ip: "localhost", ip_idx: 0, ncores: 1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081},
//			{ip: "localhost", ip_idx: 1, ncores: 1, pulse_port: 50052, log_port:50062, ctrl_port: 50072, coordinate_port: 50082}}


	// define k8s worker nodes 
//	nodes := []node{{ip: "10.10.1.2", ip_idx: -1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081, optimizer_server_port: 50091, optimizer_client_port: 50101, resources: target_resources},
//			{ip: "10.10.1.3", ip_idx: -1, pulse_port: 50052, log_port:50062, ctrl_port: 50072, coordinate_port: 50082, optimizer_server_port: 50092, optimizer_client_port: 50102, resources: target_resources},
//			{ip: "10.10.1.4", ip_idx: -1, pulse_port: 50053, log_port:50063, ctrl_port: 50073, coordinate_port: 50083, optimizer_server_port: 50093, optimizer_client_port: 50103, resources: target_resources}}

	load_pred_main(nodes)
}





