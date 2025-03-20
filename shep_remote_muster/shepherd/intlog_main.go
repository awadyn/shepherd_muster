package main

import ( "time"
	 "os"
	 "os/exec" 
//	 "fmt"
)

/**************************************/

func (intlog_s *intlog_shepherd) run_workload(m_id string) {
	l_m := intlog_s.local_musters[m_id]
	<- l_m.hb_chan

	// establish connection with remote muster and
	// get current remote node control state
//	for _, sheep := range(l_m.pasture) {
//		for _, ctrl := range(sheep.controls) {
//			l_m.request_ctrl_chan <- []string{sheep.id, ctrl.id}
//			fmt.Println("here2")
//		}
//	}
//	for _, sheep := range(l_m.pasture) {
//		for _, ctrl := range(sheep.controls) {
//			<- ctrl.ready_ctrl_chan
//			fmt.Println("hereX")
//		}
//	}
	// at this point, ctrl values are set in local muster representation
	intlog_s.init_log_files(intlog_s.intlog_musters[m_id].logs_dir)

	for iter := 0; iter < 1; iter ++ {
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

		cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
		if err := cmd.Run(); err != nil { panic(err) }

//		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=200000 --time=20")
//		cmd.Stdout = os.Stdout
//		if err := cmd.Run(); err != nil { panic(err) }
//		time.Sleep(time.Second * 5)
//		out, err := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3, 10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=200000 --time=30 | grep read | tr -s ' ' | cut -d ' ' -f9").Output()
//		if err != nil { panic(err) }

		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent=10.10.1.3 --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=256 --measure_qps=2000 --qps=100000 --time=20")
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
//		fmt.Println(out)
	}
}

func intlog_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	
	// start all management and coordination threads
	intlog_s.deploy_musters()

	for _, l_m := range(intlog_s.local_musters) {
		go intlog_s.process_logs(l_m.id)
		//go intlog_s.compute_control(l_m.id)

		go intlog_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(intlog_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





