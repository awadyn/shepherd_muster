package main

import ( "time"
	 "os"
	 "os/exec" 
	 "strconv"
	 "fmt"
)

/**************************************/

func (intlog_s *intlog_shepherd) run_workload(m_id string) {
	l_m := intlog_s.local_musters[m_id]
	<- l_m.hb_chan

	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	l_m.logs_dir = home_dir + "/" + "mustherd-logs-" + l_m.id + "/"
	err = os.Mkdir(l_m.logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	intlog_s.init_log_files(l_m.logs_dir)
  	
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

	qps_list := []int{100000, 200000, 400000, 600000, 900000}
	for _, qps := range(qps_list) {
		qps_str := strconv.Itoa(qps)
		fmt.Println("-------- QPS ", qps_str)

//		cmd = exec.Command("bash", "-c", "ssh -f awadyn@130.127.133.33 './read_rapl_rx_tx_start.sh'")
//		if err := cmd.Run(); err != nil { panic(err) }

		// TODO run workload
		cmd = exec.Command("bash", "-c", "taskset -c 10-19 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=" + qps_str + " --time=20")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }

//		cmd = exec.Command("bash", "-c", "ssh -f awadyn@130.127.133.33 './read_rapl_rx_tx.sh'")
//		if err := cmd.Run(); err != nil { panic(err) }
//		cmd = exec.Command("bash", "-c", "scp -r awadyn@130.127.133.33:~/mcd_runs_stats/ mustherd-logs-muster-130.127.133.33/stats_" + qps_str)
//		if err := cmd.Run(); err != nil { panic(err) }
		
	}

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

func intlog_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)

	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	
	// start all management and coordination threads
	intlog_s.deploy_musters()

	for _, l_m := range(intlog_s.local_musters) {
		go intlog_s.process_logs(l_m.id)
		go intlog_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(intlog_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





