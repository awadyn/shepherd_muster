package main

import ( "time"
	 "os"
	 "os/exec" 
	 "strconv"
//	 "fmt"
)

/**************************************/

func (ctrl_s *ctrl_shepherd) run_workload(m_id string) {
	c_m := ctrl_s.ctrl_musters[m_id]
	l_m := c_m.local_muster

	qps_list := []int{50000, 100000, 200000, 400000, 600000}

	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
	if err := cmd.Run(); err != nil { panic(err) }


	for _, qps := range(qps_list) {
		qps_str := strconv.Itoa(qps)

		home_dir, err := os.Getwd()
		if err != nil { panic(err) }
		l_m.logs_dir = home_dir + "/" + "intlog-logs-" + l_m.id + "-" + qps_str + "/"
		err = os.Mkdir(l_m.logs_dir, 0750)
		if err != nil && !os.IsExist(err) { panic(err) }

		ctrl_s.init_log_files(l_m.logs_dir)
	
		<- l_m.hb_chan

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

		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent=10.10.1.1 --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=256 --measure_qps=2000 --qps=" + qps_str + " --time=20")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }

		for _, sheep := range(l_m.pasture) {
			if sheep.label == "core" {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "stop", "intlogger"}
				}
			}
		}

		for _, sheep := range(l_m.pasture) {
			if sheep.label == "core" {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "close", "intlogger"}
				}
			}
		}
	}
}


func ctrl_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-ctrl"}
	s.init(nodes)

	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	ctrl_s := ctrl_shepherd{intlog_shepherd:intlog_s}
	ctrl_s.init()
	
	// start all management and coordination threads
	ctrl_s.deploy_musters()

	for _, l_m := range(ctrl_s.local_musters) {
		go ctrl_s.process_logs(l_m.id)
		go ctrl_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(ctrl_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





