package main

import ( 
	"time"
//	"strconv"
	"os"
	"os/exec" 
	"fmt"
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	<- l_m.hb_chan

	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
	if err := cmd.Run(); err != nil { panic(err) }
	cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=300000 --time=120 &")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil { panic(err) }

	for {
		select {
		case <- l_m.ready_optimize_chan:
			// run wkld with these settings
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
				}
			}
			time.Sleep(time.Second * 10)
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
			// compute reward as joules_diff
			joules_reward := reward{id:"joules", val: 0.0}
			for _, sheep := range(l_m.pasture) {		
				<- sheep.ready_ctrl_chan
				if sheep.core == 0 || sheep.core == 1 {
					diffs := bayopt_s.joules_diff[l_m.id][sheep.id]
					if len(diffs) == 0 { break }
					idx := len(diffs) - 1
					joules_reward.val += diffs[idx]
					fmt.Println("new diff: ", diffs[idx])
					fmt.Println("joules_reward: ", joules_reward.val)
				}
				// clear joules_diff map for next iteration..
				bayopt_s.joules_measure[l_m.id][sheep.id] = make([]float32, 1)
				bayopt_s.joules_diff[l_m.id][sheep.id] = make([]float32, 0)
			}
			// return reward to optimization loop
			l_m.ready_reward_chan <- reward_reply{rewards: []reward{joules_reward}}

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
	go bayopt_s.process_logs()

	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.compute_control(l_m.id)
		l_m.start_optimize_chan <- start_optimize_request{ntrials: 10}
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





