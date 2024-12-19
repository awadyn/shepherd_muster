package main

import ( "time"
	 "os"
	 "os/exec" 
)

/**************************************/

func (intlog_s *intlog_shepherd) run_workload(m_id string) {
	l_m := intlog_s.local_musters[m_id]
	<- l_m.hb_chan

	// establish connection with remote muster and
	// get current remote node control state
	for _, sheep := range(l_m.pasture) {
		for _, ctrl := range(sheep.controls) {
			l_m.request_ctrl_chan <- []string{sheep.id, ctrl.id}
			<- ctrl.ready_request_chan
		}
	}
	// at this point, ctrl values are set in local muster representation
	intlog_s.init_log_files(intlog_s.logs_dir)
//	time.Sleep(time.Second)

	for iter := 0; iter < 1; iter ++ {
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
			}
		}
//		time.Sleep(time.Second)
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "all"}
			}
		}

		cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
		if err := cmd.Run(); err != nil { panic(err) }
		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=100000 --time=30")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }

		time.Sleep(time.Second * 2)
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
			}
		}
	}
}

func intlog_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	intlog_s.init_local()
	
	// start all management and coordination threads
	intlog_s.deploy_musters()
	go intlog_s.listen_heartbeats()
	for _, l_m := range(intlog_s.local_musters) {
		go intlog_s.process_logs(l_m.id)
		//go intlog_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(intlog_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





