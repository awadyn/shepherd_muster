package main

import ( "time"
	 "os"
//	 "os/exec" 
	"fmt"
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

	time.Sleep(time.Second)

	for iter := 0; iter < 2; iter ++ {
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
				//sheep.request_log_chan <- []string{log.id, "start"}
			}
		}

		//cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=32 --measure_qps=2000 --qps=200000 --time=10")
		//cmd.Stdout = os.Stdout
		//if err := cmd.Run(); err != nil { panic(err) }
		time.Sleep(time.Second * 20)

		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
				//sheep.request_log_chan <- []string{log.id, "stop"}
			}
		}
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "first"}
				//sheep.request_log_chan <- []string{log.id, "first"}
			}
		}

		time.Sleep(time.Second)

		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "last"}
				//sheep.request_log_chan <- []string{log.id, "last"}
			}
		}

		fmt.Println("HERE")
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
				    logs_dir: home_dir + "/shepherd_intlog_logs/"}
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





