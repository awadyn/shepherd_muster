package main

import (
//	"fmt"
	"time"
//	"strconv"
	"os"
	"os/exec"
)

/**************************************/

func (lat_pred_s *latency_predictor_shepherd) run_target(m_id string) {
	lat_pred_m := lat_pred_s.lat_pred_musters[m_id]
	for {
		select {
		case opt_req := <- lat_pred_m.request_optimize_chan:
			target_ctrls := lat_pred_m.parse_optimization(opt_req)

			// set controls remotely and locally
			for _, sheep := range(lat_pred_m.pasture) {
				lat_pred_s.control(m_id, sheep.id, target_ctrls[sheep.id])
			}


			// update local log fs
			lat_pred_m.init_log_files(opt_req)
//			lat_pred_s.init_log_files(lat_pred_s.lat_pred_musters[m_id].logs_dir)


			l_m := lat_pred_m
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
				}
			}
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "all"}
				}
			}


			// run wkld, get feedback..
//			time.Sleep(time.Second * 10)
			cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
			if err := cmd.Run(); err != nil { panic(err) }
			cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=200000 --time=10")
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil { panic(err) }


			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
				}
			}
			for _, sheep := range(l_m.pasture) {
				for _, log := range(sheep.logs) {
					l_m.request_log_chan <- []string{sheep.id, log.id, "close"}
				}
			}

			latency_measure := reward{id:"latency", val: 456}
			l_m.ready_reward_chan <- reward_reply{rewards: []reward{latency_measure}}
		}
	}
}

func latency_predictor_main(nodes []node) {
	s := shepherd{id: "shepherd-latency_predictor"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	
	lat_pred_s := latency_predictor_shepherd{intlog_shepherd: intlog_s}
	lat_pred_s.init()

	// start all management and coordination threads
	lat_pred_s.deploy_musters()


	for _, l_m := range(lat_pred_s.local_musters) {
		lat_pred_s.start_optimizer(l_m.id)
		lat_pred_s.run_target(l_m.id)
	}


	time.Sleep(exp_timeout)

	
	for _, l_m := range(lat_pred_s.local_musters) {
		lat_pred_s.stop_optimizer(l_m.id)
	}


	lat_pred_s.cleanup()
}







