package main

import ( "time"
	 "os"
	 "os/exec" 
	 "fmt"
)

/**************************************/

func (stats_s *stats_shepherd) run_workload(m_id string) {
	l_m := stats_s.local_musters[m_id]
	s_m := stats_s.stats_musters[m_id]
	fmt.Println("local:", l_m.logs_dir)
	fmt.Println("intlog:", s_m.intlog_muster.logs_dir)
	fmt.Println("stats:", s_m.logs_dir)
	<- l_m.hb_chan

	stats_s.init_log_files(l_m.logs_dir)

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
	}
}


func stats_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-stats"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	stats_s := stats_shepherd{shepherd:s}
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





