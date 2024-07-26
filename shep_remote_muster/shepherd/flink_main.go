package main

import ( "time"
)

/**************************************/


func (flink_s *flink_shepherd) run_workload(m_id string) {
	l_m := flink_s.local_musters[m_id]
	<- l_m.hb_chan

	for iter := 0; iter < 2; iter ++ {
		for sheep_id, sheep := range(l_m.pasture) {
			for log_id, _ := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep_id, log_id, "start"}
			}
		}	

		time.Sleep(time.Second * 10)

		for sheep_id, sheep := range(l_m.pasture) {
			for log_id, _ := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep_id, log_id, "stop"}
			}
		}

		for sheep_id, sheep := range(l_m.pasture) {
			for log_id, _ := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep_id, log_id, "first"}
			}
		}
		for sheep_id, sheep := range(l_m.pasture) {
			for log_id, _ := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep_id, log_id, "last"}
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func flink_main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.10.1.1", ncores: 1, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071, coordinate_port: 50081},
			{ip: "10.10.1.3", ncores: 1, pulse_port: 50053, log_sync_port:50063, ctrl_port: 50073, coordinate_port: 50083}}

	// initialize generic shepherd
	s := shepherd{id: "sheperd-flink"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	flink_s := flink_shepherd{shepherd:s}
	flink_s.init()
	flink_s.show()
	flink_s.deploy_musters()
	go flink_s.listen_heartbeats()
//	go flink_s.process_logs()
//	go flink_s.compute_control()
//	for _, l_m := range(flink_s.local_musters) {
//		go flink_s.run_workload(l_m.id)
//	}

	time.Sleep(exp_timeout)

	/* close any open output files */
	for _, l_m := range(flink_s.local_musters) {
		for sheep_id, _ := range(l_m.pasture) {
			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
		}
	}
}





