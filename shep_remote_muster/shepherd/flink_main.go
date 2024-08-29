package main

import ( "time"
//	 "fmt"
)

/**************************************/
//func (m *muster) send_log_cmd(cmd string) {
//	for sheep_id, sheep := range(m.pasture) {
//		for log_id, _ := range(sheep.logs) {
//			m.request_log_chan <- []string{sheep_id, log_id, cmd}
//		}
//	}	
//}

//func (flink_s *flink_shepherd) run_workload(m_id string) {
//	m := flink_s.musters[m_id]
//	<- m.hb_chan
//
//	for iter := 0; iter < 2; iter ++ {
//		m.send_log_cmd("start")
//		time.Sleep(time.Second * 10)
//		m.send_log_cmd("stop")
//		m.send_log_cmd("first")
//		m.send_log_cmd("last")
//		time.Sleep(time.Second)
//	}
//}

func flink_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-flink"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	flink_s := flink_shepherd{shepherd:s}
	flink_s.init()
	flink_s.deploy_musters()
	go flink_s.listen_heartbeats()
//	go flink_s.process_logs()
//	go flink_s.compute_control()
//	for _, l_m := range(flink_s.local_musters) {
//		go flink_s.run_workload(l_m.id)
//	}

	time.Sleep(exp_timeout)

//	/* close any open output files */
//	for _, l_m := range(flink_s.local_musters) {
//		for sheep_id, _ := range(l_m.pasture) {
//			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
//		}
//	}
}





