package main

import ( 
	"fmt"
	"time"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)



func load_pred_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-load-pred"}
	s.init(nodes)
	
	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	load_pred_s := load_pred_shepherd{intlog_shepherd:intlog_s}
	load_pred_s.init()

	if optimize_on {
		optimizer_server_ports := []uint64{50091}
		optimizer_client_ports := []uint64{50101}
		load_pred_s.init_optimizers(optimizer_server_ports, optimizer_client_ports)
	}

	load_pred_s.show()

	// start all management and coordination threads
	load_pred_s.deploy_musters()

	for _, l_m := range(load_pred_s.local_musters) {
		go load_pred_s.process_logs(l_m.id)
		go load_pred_s.start_exp(l_m.id)
	}
//	go load_pred_s.process_control()

	if optimize_on {
		for id, optimizer := range(load_pred_s.optimizers) {
			start_args := make([]*anypb.Any, 0)
			data := wrapperspb.String("/users/awadyn/shepherd_muster/shep_remote_muster/mustherd-logs-muster-10.10.1.2/")
			arg, _ := anypb.New(data)
			start_args = append(start_args, arg)
			optimizer.start_optimize_chan <- start_optimize_request{args: start_args}
			done := <- optimizer.ready_optimize_chan
			if done {
				fmt.Printf("** ** ** OPTIMIZER %v READY\n", id)
			} else {
				fmt.Printf("** ** ** ERROR ** ** ** OPTIMIZER %v READY\n", id)
			}

		}
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(load_pred_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}




