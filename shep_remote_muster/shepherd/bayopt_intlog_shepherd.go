package main

import (
//	"os"
	"fmt"
	"strconv"
//	"slices"
//	"sort"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

//type bayopt_intlog_muster struct {
//	intlog_muster
//}

type bayopt_intlog_shepherd struct {
	shepherd
	bayopt_intlog_musters map[string]*bayopt_intlog_muster 
}


func (bayopt_s *bayopt_intlog_shepherd) init() {
	bayopt_s.bayopt_intlog_musters = make(map[string]*bayopt_intlog_muster)
	for _, l_m := range(bayopt_s.local_musters) {
		intlog_m := intlog_muster{local_muster: *l_m}
		intlog_m.init()

		bayopt_m := bayopt_intlog_muster{intlog_muster: intlog_m}
		bayopt_m.init()

		bayopt_s.bayopt_intlog_musters[bayopt_m.id] = &bayopt_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (bayopt_s bayopt_intlog_shepherd) process_logs(m_id string) {
	l_m := bayopt_s.bayopt_intlog_musters[m_id]
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				sheep := sheep
				log := log
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id) }
				select {
				case log.ready_process_chan <- true:
				default:
				}
				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }

			} ()
		}
	}
}


func (l_m *bayopt_intlog_muster) parse_optimization(opt_settings []optimize_setting) map[string]map[string]uint64 {
	target_ctrls := make(map[string]map[string]uint64)
	for sheep_id, _ := range(l_m.pasture) {
		target_ctrls[sheep_id] = make(map[string]uint64)
	}

	for _, opt_setting := range(opt_settings) {
		switch {
		case opt_setting.knob == "itr-delay":
			for _, sheep := range(l_m.pasture) {
				if sheep.label == "node" {
					target_ctrls[sheep.id]["itr-ctrl-" + sheep.label + "-" + strconv.Itoa(int(sheep.index)) + "-" + l_m.ip] = opt_setting.val 
					break
				}
			}
		case opt_setting.knob == "dvfs":
			for _, sheep := range(l_m.pasture) {
				if sheep.label == "core" {
					target_ctrls[sheep.id]["dvfs-ctrl-" + sheep.label + "-" + strconv.Itoa(int(sheep.index)) + "-" + l_m.ip] = opt_setting.val
				}
			}
		default:
			fmt.Println("****** Unimplemented optimization setting: ", opt_setting.knob)
		}
	}
	return target_ctrls
}

func (bayopt_s bayopt_intlog_shepherd) process_control(m_id string) {
	l_m := bayopt_s.bayopt_intlog_musters[m_id]
	for {
		select {
		case opt_req := <- l_m.request_optimize_chan:
			opt_settings := opt_req.settings

			target_ctrls := l_m.parse_optimization(opt_settings)
			l_m.request_ctrl_chan <- target_ctrls
			<- l_m.done_request_chan

			select {
			case l_m.ready_optimize_chan <- true:
			default:
			}
		}
	}		
}





