package main

//import "fmt"

/**************************************/

var specialize_on = true
var optimize_on bool = false
var debug bool = true

func main() {
	// assume a list of resources per-node is known apriori
	target_resources := make([]resource, 0)
	var i uint8
	for i = 0; i < 16; i++ {
		target_resources = append(target_resources, resource{label: "core", index: i})
	}
	target_resources = append(target_resources, resource{label: "node", index: 0})

	// assume that a list of nodes is known apriori
//	nodes := []node{{ip: "10.10.1.2", ip_idx: -1, ncores: 16, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081, optimizer_server_port: 50091, optimizer_client_port: 50101}}
	nodes := []node{{ip: "10.10.1.2", ip_idx: -1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081, optimizer_server_port: 50091, optimizer_client_port: 50101, resources: target_resources}}
//	nodes := []node{{ip: "localhost", ip_idx: -1, ncores: 16, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081}}
//	nodes := []node{{ip: "localhost", ip_idx: 0, ncores: 1, pulse_port: 50051, log_port:50061, ctrl_port: 50071, coordinate_port: 50081},
//			{ip: "localhost", ip_idx: 1, ncores: 1, pulse_port: 50052, log_port:50062, ctrl_port: 50072, coordinate_port: 50082}}

//	args := make(map[string]map[string]string)
//	args[nodes[0].ip] := make(map[string]string)
//	args[nodes[0].ip]["opt_type"] = "bayopt"
//	args[nodes[0].ip]["num_trials"] = "30"


	ctrl_main(nodes)
}





