package main

import ( "time"
//	 "os"
//	 "os/exec" 
//	 "strconv"
//	 "fmt"
)

/**************************************/

//var medians []int = []int{11035, 19450, 37076, 74224, 112588}
//var slopes []float64 = []float64{5.942, 5.673, 5.384, 5.213}
//var intercepts []float64= []float64{-15569.97, -10339.85, 382.816, 13070.288}

//func (stats_s *stats_shepherd) run_workload(m_id string) {
////	l_m := stats_s.local_musters[m_id]
//	s_m := stats_s.stats_musters[m_id]
//	l_m := s_m.local_muster
//	<- l_m.hb_chan
//
//	stats_s.init_log_files(l_m.logs_dir)
//
//	qps_list := []int{50000, 100000, 200000, 400000, 600000}
//
//	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
//	if err := cmd.Run(); err != nil { panic(err) }
//
//	for _, sheep := range(l_m.pasture) {
//		if sheep.label == "core" {
//			for _, log := range(sheep.logs) {
//				l_m.request_log_chan <- []string{sheep.id, log.id, "start", "intlogger"}
//			}
//		}
//	}
//	for _, sheep := range(l_m.pasture) {
//		if sheep.label == "core" {
//			for _, log := range(sheep.logs) {
//				l_m.request_log_chan <- []string{sheep.id, log.id, "all", "intlogger"}
//			}
//		}
//	}
//
//	for _, qps := range(qps_list) {
//		qps_str := strconv.Itoa(qps)
//
////		stats_s.init_log_files(l_m.logs_dir)
////		time.Sleep(time.Second)
//
////		for _, sheep := range(l_m.pasture) {
////			if sheep.label == "core" {
////				for _, log := range(sheep.logs) {
////					l_m.request_log_chan <- []string{sheep.id, log.id, "start", "intlogger"}
////				}
////			}
////		}
////		for _, sheep := range(l_m.pasture) {
////			if sheep.label == "core" {
////				for _, log := range(sheep.logs) {
////					l_m.request_log_chan <- []string{sheep.id, log.id, "all", "intlogger"}
////				}
////			}
////		}
//
//		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent=10.10.1.1 --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=256 --measure_qps=2000 --qps=" + qps_str + " --time=20")
//		cmd.Stdout = os.Stdout
//		if err := cmd.Run(); err != nil { panic(err) }
//
////		for _, sheep := range(l_m.pasture) {
////			if sheep.label == "core" {
////				for _, log := range(sheep.logs) {
////					l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
////				}
////			}
////		}
////		time.Sleep(time.Second * 1)
////
////		for _, sheep := range(l_m.pasture) {
////			if sheep.label == "core" {
////				for _, log := range(sheep.logs) {
////					l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
////				}
////			}
////		}
//
//		time.Sleep(time.Second)
//
//		fmt.Println("MEDIAN: ", s_m.rx_bytes_median)
//		fmt.Println("Quessing QPS... ")
//		var i int
//		for j, med := range(medians) {
//			if s_m.rx_bytes_median <= med { 
//				i = j
//				break 
//			}
//		}
//		m := slopes[i-1]
//		b := intercepts[i-1]
//		guess := int(m) * s_m.rx_bytes_median + int(b)
//		fmt.Println("qps real: ", qps, "qps guess: ", guess)
//		
//
//		stats_s.rx_bytes_medians = append(stats_s.rx_bytes_medians, s_m.rx_bytes_median)
//		time.Sleep(time.Second)
//	}
//
//	for _, sheep := range(l_m.pasture) {
//		if sheep.label == "core" {
//			for _, log := range(sheep.logs) {
//				l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
//			}
//		}
//	}
//	for _, sheep := range(l_m.pasture) {
//		if sheep.label == "core" {
//			for _, log := range(sheep.logs) {
//				l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
//			}
//		}
//	}
//
//	fmt.Println(stats_s.rx_bytes_medians)
//}


func stats_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-stats"}
	s.init(nodes)

	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	stats_s := stats_shepherd{intlog_shepherd:intlog_s}
	stats_s.init()
	
	// start all management and coordination threads
	stats_s.deploy_musters()

	for _, l_m := range(stats_s.local_musters) {
	
		go stats_s.process_logs(l_m.id)
		//go intlog_s.process_logs(l_m.id)

		go intlog_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(stats_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





