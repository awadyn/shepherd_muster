package main

import ( "time"
	 "os"
	 "os/exec" 
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
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
	bayopt_s.init_log_files(bayopt_s.logs_dir)

//	time.Sleep(time.Second)

	for iter := 0; iter < 1; iter ++ {
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
			}
		}

		cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
		if err := cmd.Run(); err != nil { panic(err) }
		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=100000 --time=30")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }
		time.Sleep(time.Second)

		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
			}
		}
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "first"}
			}
		}

		time.Sleep(time.Second)

		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "last"}
			}
		}

		time.Sleep(time.Second)
	}
}

func bayopt_main(nodes []node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	// initialize generic shepherd
	s := shepherd{id: "sheperd-bayopt"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	bayopt_s := bayopt_shepherd{shepherd:s,
				    logs_dir: home_dir + "/shepherd_bayopt_logs/"}
	bayopt_s.init()
	bayopt_s.init_local()
	
	// start all management and coordination threads
	bayopt_s.deploy_musters()
	go bayopt_s.listen_heartbeats()
	go bayopt_s.process_logs()
//	go bayopt_s.compute_control()
	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)

	for _, l_m := range(bayopt_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





