package main

import ( "time"
	 "os"
	 "os/exec" 
//	 "strconv"
	 "fmt"
//	 "sort"
)

/**************************************/
// OFFLINE PHASE
//var qpses []int = make([]int, 0)
//var medians []int = make([]int, 0)
//var medians map[int]int

// ONLINE PHASE
//var qpses []int = []int{50000, 100000, 200000, 400000, 600000, 750000, 900000}
//var qpses []int = []int{100000, 200000, 400000, 600000, 900000}
//var medians []int = []int{17576, 33329, 62459, 86988, 114710}

func (stats_s *stats_shepherd) run_workload(m_id string) {
	l_m := stats_s.stats_musters[m_id]
	<- l_m.hb_chan

	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	l_m.logs_dir = home_dir + "/" + "mustherd-logs-" + l_m.id + "/"
	err = os.Mkdir(l_m.logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	stats_s.init_log_files(l_m.logs_dir)
  
	cmd_setup_mutilate := "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value;"
	//cmd_setup_mutilate += "echo > ~/all_mutilate_output.txt;"
	//cmd_setup_mutilate += "ssh -f " + l_m.ip + " './read_rapl_start.sh';"
	cmd := exec.Command("bash", "-c", cmd_setup_mutilate)
	if err := cmd.Run(); err != nil { 
		fmt.Printf("\033[31;1m****** ERROR: MUTILATE -- failed to load database: %v\n\033[0m", err)
		panic(err) 
	}

	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start", "intlogger"}
			}
		}
	}
	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "all", "intlogger"}
			}
		}
	}

	time.Sleep(time.Second * exp_timeout - 5)

//	time.Sleep(time.Second)
//
//	// ONLINE (TESTING) PHASE
//	//qps_list := []int{400000, 900000, 900000, 600000, 200000, 400000, 600000, 100000, 600000, 200000, 100000, 400000, 900000}
//	qps_list := []int{400000, 900000, 600000, 200000, 100000, 400000}
//	for _, qps := range(qps_list) {
//		qps_str := strconv.Itoa(qps)
//		fmt.Println("--------------------------------------- QPS ", qps_str, " ---------------------------------------------------------")
//
//		// TODO run workload
//		cmd = exec.Command("bash", "-c", "taskset -c 10-19 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=10 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=" + qps_str + " --time=20")
//		cmd.Stdout = os.Stdout
//		if err := cmd.Run(); err != nil { panic(err) }
//
////		cmd_measure := "ssh -f " + l_m.ip + " './read_rapl.sh';"
////		cmd = exec.Command("bash", "-c", cmd_measure)
////		if err := cmd.Run(); err != nil { panic(err) }
//
//		l_m.rx_bytes_medians = make([]int, 0)
//	}
//
//	// OFFLINE (LEARNING) PHASE
///*	
//	// generating rx_bytes medians to itrd to qps mappings
//	qps_list := []int{50000, 100000, 200000, 400000, 600000, 750000, 900000}
//	itrd_list := []int{1, 5, 10, 20, 50, 100, 200, 300}
//	for _, itrd := range(itrd_list) {
//		fmt.Println("--------------------------------------- ITRD ", itrd, " ---------------------------------------------------------")
//		medians[itrd] := make([]int, 0)
//		itrd_str := str(itrd)
//		for _, qps := range(qps_list) {
//			stats_s.rx_bytes_medians[qps] = make([]int, 0)
//			fmt.Println("--------------------------------------- QPS ", qps, " ---------------------------------------------------------")
//	
//			cmd = exec.Command("bash", "-c", "~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=8 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=" + qps_str + " --time=20 >> ~/all_mutilate_output.txt")
//			cmd.Stdout = os.Stdout
//			if err := cmd.Run(); err != nil { panic(err) }
//
//      			stats_s.rx_bytes_medians[qps] = l_m.rx_bytes_medians
//      			l_m.rx_bytes_medians = make([]int, 0)
//
//			len_medians := len(stats_s.rx_bytes_medians[qps])
//			median := stats_s.rx_bytes_medians[qps][len_medians/2]
//			medians[itrd] = append(medians[itrd], median)
//		}
//	}
//*/
//
//	time.Sleep(time.Second)
	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
			}
		}
	}
	/*
	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
			}
		}
	}
	*/
}

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
		go stats_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(stats_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}

