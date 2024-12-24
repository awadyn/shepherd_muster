package main

import ( 
	"time"
	"strconv"
	"os"
	"os/exec" 
	"fmt"
	"math"
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	<- l_m.hb_chan

	for {
		select {
		case <- l_m.ready_optimize_chan:
			// run wkld with these settings
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
				}
			}
			cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
			if err := cmd.Run(); err != nil { panic(err) }
			out, err := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=32 --measure_qps=2000 --qps=600000 --time=30 | grep read | tr -s ' ' | cut -d ' ' -f9").Output()
			cmd.Stdout = os.Stdout
			if err != nil { panic(err) }
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
				}
			}

			// get first and last logs, computing joules_diff while processing logs
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "first"}
				}
			}
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "last"}
				}
			}
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "close"}
				}
			}

			//reward_rep := bayopt_s.bayopt_musters[l_m.id].compute_energy_reward()

			// compute reward as joules_diff
			joules_reward := reward{id:"joules", val: 0.0}
			var lat_val float64
			if lat_val, err = strconv.ParseFloat(string(out[0:len(out)-2]), 64); err != nil { panic(err) }
			latency_reward := reward{id: "latency", val: float32(lat_val)}
			var joules_min_0 float32
			var joules_max_0 float32
			var joules_diff_0 float32
			var joules_min_1 float32
			var joules_max_1 float32
			var joules_diff_1 float32
			timestamp_min_0 := float32(math.MaxFloat32)
			timestamp_max_0 := float32(math.SmallestNonzeroFloat32)
			timestamp_min_1 := float32(math.MaxFloat32)
			timestamp_max_1 := float32(math.SmallestNonzeroFloat32)

			for _, sheep := range(l_m.pasture) {		
				<- sheep.ready_ctrl_chan
				switch {
				case sheep.core <= 7:
					if sheep.perf_data["timestamp_measure"][0] < timestamp_min_0 {
						timestamp_min_0 = sheep.perf_data["timestamp_measure"][0]
						joules_min_0 = sheep.perf_data["joules_measure"][1]
					}
					if sheep.perf_data["timestamp_measure"][1] > timestamp_max_0 {
						timestamp_max_0 = sheep.perf_data["timestamp_measure"][1]
						joules_max_0 = sheep.perf_data["joules_measure"][2]
					}
					if joules_max_0 < joules_min_0 {
						joules_diff_0 = float32(math.Pow(2, 32) * 0.000061) - joules_min_0 + joules_max_0
					} else {
						joules_diff_0 = joules_max_0 - joules_min_0
					}
				case sheep.core > 7:
					if sheep.perf_data["timestamp_measure"][0] < timestamp_min_1 {
						timestamp_min_1 = sheep.perf_data["timestamp_measure"][0]
						joules_min_1 = sheep.perf_data["joules_measure"][1]
					}
					if sheep.perf_data["timestamp_measure"][1] > timestamp_max_1 {
						timestamp_max_1 = sheep.perf_data["timestamp_measure"][1]
						joules_max_1 = sheep.perf_data["joules_measure"][2]
					}
					if joules_max_1 < joules_min_1 {
						joules_diff_1 = float32(math.Pow(2, 32) * 0.000061) - joules_min_1 + joules_max_1
					} else {
						joules_diff_1 = joules_max_1 - joules_min_1
					}
				}
				// clear joules_diff map for next iteration..
				sheep.perf_data["joules_measure"] = make([]float32, 1)
				sheep.perf_data["joules_diff"] = make([]float32, 0)
				sheep.perf_data["timestamp_measure"] = make([]float32, 0)
			}

			fmt.Printf("\npackage 0: min %v - max %v\n", joules_min_0, joules_max_0)
			fmt.Printf("timestamp: min %v - max %v\n", timestamp_min_0, timestamp_max_0)
			fmt.Printf("\npackage 1: min %v - max %v\n", joules_min_1, joules_max_1)
			fmt.Printf("timestamp: min %v - max %v\n\n", timestamp_min_1, timestamp_max_1)

			joules_reward.val = joules_diff_0 + joules_diff_1

			// return reward to optimization loop
			l_m.ready_reward_chan <- reward_reply{rewards: []reward{joules_reward, latency_reward}}
			// now, optimization loop is expected to send another optmization setting request, which will repeat the above process
		}
	}
}

func bayopt_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-bayopt"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	bayopt_s := bayopt_shepherd{shepherd:s}
	bayopt_s.init()
	bayopt_s.init_local()
	
	// start all management and coordination threads
	bayopt_s.deploy_musters()
	go bayopt_s.listen_heartbeats()

	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.process_logs(l_m.id)
		go bayopt_s.compute_control(l_m.id)

		l_m.start_optimize_chan <- start_optimize_request{ntrials: 40}
		done := <- l_m.ready_optimize_chan
		if done {
			//running workload with optimization ready..
			go bayopt_s.run_workload(l_m.id)
		}
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(bayopt_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { defer f.Close() }
		}
	}
}





