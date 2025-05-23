package main

import ( "time"
	 "os"
	 "os/exec" 
	 "strconv"
	 "fmt"
//	 "sort"
)

/**************************************/
//var qpses []int = make([]int, 0)
//var medians []int = make([]int, 0)
var qpses []int = []int{100000, 200000, 400000, 600000, 900000}
var medians []int = []int{17576, 33329, 62459, 86988, 114710}
var opt_dvfs []string = []string{"0xc00", "0x1100", "0x1900", "0x1500", "0x1700"}
var opt_itrd []string = []string{"50", "200", "150", "100", "100"}

func (stats_s *stats_shepherd) run_workload(m_id string) {
	l_m := stats_s.stats_musters[m_id]
	<- l_m.hb_chan

	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	l_m.logs_dir = home_dir + "/" + "mustherd-logs-" + l_m.id + "/"
	err = os.Mkdir(l_m.logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	stats_s.init_log_files(l_m.logs_dir)
  	
	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
	if err := cmd.Run(); err != nil { panic(err) }
	time.Sleep(time.Second)

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
	time.Sleep(time.Second)

	//qps_list := []int{100000, 200000, 400000, 600000, 900000}
	cmd = exec.Command("bash", "-c", "echo > all_mutilate_output.txt")
	if err := cmd.Run(); err != nil { panic(err) }
	
	cmd = exec.Command("bash", "-c", "ssh -f awadyn@130.127.133.33 './read_rapl_rx_tx_start.sh'")
	if err := cmd.Run(); err != nil { panic(err) }

	qps_list := []int{400000, 900000, 900000, 600000, 200000, 400000, 600000, 100000, 600000, 200000, 100000, 400000, 900000}
	for _, qps := range(qps_list) {
		qps_str := strconv.Itoa(qps)
		stats_s.rx_bytes_medians[qps] = make([]int, 0)
		fmt.Println("-------- QPS ", qps_str, " ---------------------------------------------------------")

		// TODO run workload
		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=512 --measure_qps=2000 --qps=" + qps_str + " --time=30 >> all_mutilate_output.txt")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }

		cmd = exec.Command("bash", "-c", "ssh -f awadyn@130.127.133.33 './read_rapl_rx_tx.sh'")
		if err := cmd.Run(); err != nil { panic(err) }
//		cmd = exec.Command("bash", "-c", "scp -r awadyn@130.127.133.33:~/mcd_runs_stats/ mustherd-logs-muster-130.127.133.33/stats_" + qps_str)
//		if err := cmd.Run(); err != nil { panic(err) }
		
		fmt.Println(l_m.rx_bytes_medians)
		stats_s.rx_bytes_medians[qps] = l_m.rx_bytes_medians
		fmt.Println(stats_s.rx_bytes_medians)
		l_m.rx_bytes_medians = make([]int, 0)

//		qpses = append(qpses, qps)
//		sort.Ints(stats_s.rx_bytes_medians[qps])
//		len_medians := len(stats_s.rx_bytes_medians[qps])
//		median := stats_s.rx_bytes_medians[qps][len_medians/2]
//		medians = append(medians, median)
	}
//	fmt.Println(stats_s.rx_bytes_medians)
//	fmt.Println(qpses)
//	fmt.Println(medians)

	time.Sleep(time.Second)
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

