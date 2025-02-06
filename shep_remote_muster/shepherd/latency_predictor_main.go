package main

import (
	"fmt"
	"time"
	"strconv"
)

/**************************************/
type latency_predictor_muster struct {
	intlog_muster
}

type latency_predictor_shepherd struct {
	intlog_shepherd 
}

func (lat_pred_m *latency_predictor_muster) init() {
	for sheep_id, sheep := range(lat_pred_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		switch {
		case label == "core":
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + lat_pred_m.ip, n_ip: lat_pred_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
			lat_pred_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + lat_pred_m.ip, n_ip: lat_pred_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			lat_pred_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
		default:
		}
	}
}

func (lat_pred_s *latency_predictor_shepherd) init() {
	for _, intlog_m := range(lat_pred_s.intlog_musters) {
		lat_pred_m := latency_predictor_muster{intlog_muster: *intlog_m}
		lat_pred_m.init()
		lat_pred_m.show()
	}
}

func (intlog_s *intlog_shepherd) run_target(m_id string) {
	l_m := intlog_s.local_musters[m_id]
	for {
		select {
		case opt_req := <- l_m.request_optimize_chan:
			// parse optimize settings to target controls 
			// as-per specialization: this shepherd expects 1 dvfs setting for all cores 
			//			  and 1 itr-delay setting for the full node

			// init map of sheep --> ctrl --> ctrl-val
			target_ctrls := make(map[string]map[string]uint64)
			for sheep_id, _ := range(l_m.pasture) {
				target_ctrls[sheep_id] = make(map[string]uint64)
			}

			skip_sheep := false
			for _, opt_setting := range(opt_req.settings) {
				for _, sheep := range(l_m.pasture) {
					if skip_sheep { break }
					switch {
					case opt_setting.knob == "itr-delay":
						target_ctrls[sheep.id]["itr-ctrl-" + l_m.ip] = opt_setting.val
						skip_sheep = true
					case opt_setting.knob == "dvfs":
						c_str := strconv.Itoa(int(sheep.core))
						target_ctrls[sheep.id]["dvfs-ctrl-" + c_str + "-" + l_m.ip] = opt_setting.val
					default:
						fmt.Println("****** Unimplemented optimization control setting: ", opt_setting)
					}
				}
				skip_sheep = false
			}

			// set controls remotely and locally
			for _, sheep := range(l_m.pasture) {
				intlog_s.control(l_m.id, sheep.id, target_ctrls[sheep.id])
			}

			// update local log fs
			intlog_s.init_log_files(intlog_s.intlog_musters[m_id].logs_dir)

			// run wkld, get feedback..

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
	intlog_s.deploy_musters()


	intlog_s.start_optimizer()

	for _, l_m := range(intlog_s.local_musters) {
		intlog_s.run_target(l_m.id)
	}


	time.Sleep(exp_timeout)

	
	intlog_s.stop_optimizer()


	intlog_s.cleanup()

}







