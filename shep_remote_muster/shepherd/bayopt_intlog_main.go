package main

import ( "time"
	 "os"
//	 "os/exec" 
//	 "strconv"
	 "fmt"
)

/**************************************/

func (bayopt_s *bayopt_intlog_shepherd) run_workload(m_id string) {
	b_m := bayopt_s.bayopt_intlog_musters[m_id]
	l_m := b_m.local_muster
	<- l_m.hb_chan

	qps_list := []int{200000, 400000, 600000}
//	qps_list := []int{50000, 100000, 200000, 400000, 600000, 200000, 500000, 750000, 900000, 600000, 300000}


//	var dvfs_val uint64 = 0x1900
//	var itrd_val uint64 = 400
//	// set ctrls
//	for _, sheep := range(l_m.pasture) {
//		start_ctrls := make(map[string]uint64)
//		index := strconv.Itoa(int(sheep.index))
//		if sheep.label == "core" {
//			ctrl_dvfs_id := "dvfs-ctrl-" + sheep.label + "-" + index  + "-" + l_m.ip
//			start_ctrls[ctrl_dvfs_id] = dvfs_val
//		}
//		if sheep.label == "node" {
//			ctrl_itr_id := "itr-ctrl-" + sheep.label + "-" + index + "-" + l_m.ip
//			start_ctrls[ctrl_itr_id] = itrd_val
//		}
//		bayopt_s.control(l_m.id, sheep.id, start_ctrls)
//	}


//	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
//	if err := cmd.Run(); err != nil { panic(err) }
//
//	time.Sleep(time.Second)

	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	l_m.logs_dir = home_dir + "/" + "mustherd-logs-" + l_m.id + "/"
	err = os.Mkdir(l_m.logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	bayopt_s.init_log_files(l_m.logs_dir)
  	time.Sleep(time.Second)

	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start", "intlogger"}
			}
		}
	}
	time.Sleep(time.Second*2)

	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "all", "intlogger"}
			}
		}
	}
	time.Sleep(time.Second*2)

	for _, qps := range(qps_list) {
//		qps_str := strconv.Itoa(qps)
		fmt.Println(qps)

//		l_m.logs_dir = home_dir + "/" + "mustherd-logs-" + l_m.id + "-" + qps_str + "/"
//		err = os.Mkdir(l_m.logs_dir, 0750)
//		if err != nil && !os.IsExist(err) { panic(err) }
//		bayopt_s.init_log_files(l_m.logs_dir)
//
//		for _, sheep := range(l_m.pasture) {
//			if sheep.label == "core" {
//				for _, log := range(sheep.logs) {
//					l_m.request_log_chan <- []string{sheep.id, log.id, "start", "intlogger"}
//				}
//			}
//		}
//		for _, sheep := range(l_m.pasture) {
//			if sheep.label == "core" {
//				for _, log := range(sheep.logs) {
//					l_m.request_log_chan <- []string{sheep.id, log.id, "all", "intlogger"}
//				}
//			}
//		}

//		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=256 --measure_qps=2000 --qps=" + qps_str + " --time=20")
//		cmd.Stdout = os.Stdout
//		if err := cmd.Run(); err != nil { panic(err) }
		time.Sleep(time.Second * 5)

//		for _, sheep := range(l_m.pasture) {
//			if sheep.label == "core" {
//				for _, log := range(sheep.logs) {
//					l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
//				}
//			}
//		}
//		for _, sheep := range(l_m.pasture) {
//			if sheep.label == "core" {
//				for _, log := range(sheep.logs) {
//					l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
//				}
//			}
//		}
	}

	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
			}
		}
	}
	time.Sleep(time.Second*2)

	for _, sheep := range(l_m.pasture) {
		if sheep.label == "core" {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
			}
		}
	}
}


func bayopt_intlog_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-bayopt-intlog"}
	s.init(nodes)

	bayopt_s := bayopt_intlog_shepherd{shepherd:s}
	bayopt_s.init()
	
	// start all management and coordination threads
	bayopt_s.deploy_musters()

	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.process_logs(l_m.id)
		go bayopt_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(bayopt_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





