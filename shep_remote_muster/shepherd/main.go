package main

import ( "time" )

/**************************************/

func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.10.1.2", ncores: 16, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071}}
	//nodes := []node{{ip: "128.110.96.54", ncores: 16, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071}}

	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	// for each muster, start pulse + log + control threads for a total
	// of num_musters * [1(pulse client) + 2(log server + coordinator) + 1(ctrl client)]
	// = 4 * num_musters
	intlog_s.deploy_musters()

	// 1 thread listening for muster pulses
	go intlog_s.listen_heartbeats()

	// 1 thread managing process signals + (0 <= threads <= muster.ncores) 
	// doing actual log processing 
	go intlog_s.process_logs()

//	go intlog_s.compute_control()

	time.Sleep(exp_timeout)
	for _, l_m := range(intlog_s.local_musters) {
		for sheep_id, _ := range(l_m.pasture) {
			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
		}
	}
}





