package main

import ( "time"
//	 "os"
//	 "os/exec" 
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	<- l_m.hb_chan

	for iter := 0; iter < 2; iter ++ {
		for sheep_id, sheep := range(l_m.pasture) {
			for log_id, _ := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep_id, log_id, "start"}
			}
		}	
//		cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=32 --measure_qps=2000 --qps=200000 --time=10")
//		cmd.Stdout = os.Stdout
//		if err := cmd.Run(); err != nil { panic(err) }
		time.Sleep(time.Second * 20)
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

func bayopt_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-bayopt"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	bayopt_s := bayopt_shepherd{shepherd:s}
	bayopt_s.init()
	
	// start all management and coordination threads
	bayopt_s.deploy_musters()
	go bayopt_s.listen_heartbeats()
	go bayopt_s.process_logs()
	go bayopt_s.compute_control()
	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)

	for _, l_m := range(bayopt_s.local_musters) {
		for sheep_id, _ := range(l_m.pasture) {
			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
		}
	}
}





