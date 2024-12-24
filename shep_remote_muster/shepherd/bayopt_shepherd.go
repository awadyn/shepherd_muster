package main

import (
	"fmt"
	"strconv"
	"slices"
	"os"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

var joules_idx int
var timestamp_idx int 

type bayopt_muster struct {
	local_muster
	ixgbe_metrics []string
	buff_max_size uint64
}

type bayopt_shepherd struct {
	shepherd
	bayopt_musters map[string]*bayopt_muster 
	ixgbe_metrics []string
	buff_max_size uint64
}


/* 
   This function initializes a specialized shepherd for energy-and-performance 
   management. Each muster under shepherd supervision produces energy and
   performance logs and applies control changes to energy and performance 
   settings - i.e. interrupt delay and dynamic-voltage-frequency-scaling - 
   upon shepherd decision-making for each core under muster supervision.
*/
func (bayopt_s *bayopt_shepherd) init() {
	bayopt_s.ixgbe_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
	                                "instructions", "cycles", "ref_cycles", "llc_miss",
	                                "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	bayopt_s.buff_max_size = 1
	bayopt_s.bayopt_musters = make(map[string]*bayopt_muster)
	joules_idx = slices.Index(bayopt_s.ixgbe_metrics, "joules")
	timestamp_idx = slices.Index(bayopt_s.ixgbe_metrics, "timestamp")

	for _, l_m := range(bayopt_s.local_musters) {
		bayopt_m := bayopt_muster{local_muster: *l_m}
		bayopt_m.init()
		bayopt_s.bayopt_musters[bayopt_m.id] = &bayopt_m
	}
}

func (bayopt_s *bayopt_shepherd) init_local() {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	for _, bayopt_m := range(bayopt_s.bayopt_musters) {
		logs_dir := home_dir + "/" + bayopt_m.id + "-bayopt-logs/"
		bayopt_m.init_local(logs_dir)
		for _, sheep := range(bayopt_m.pasture) {
			sheep.perf_data["joules_measure"] = make([]float32, 1)
			sheep.perf_data["joules_diff"] = make([]float32, 0)
			sheep.perf_data["timestamp_measure"] = make([]float32, 0)
			for _, log := range(sheep.logs) {
				log.ready_process_chan <- true
				log.ready_request_chan <- true
				log.ready_buff_chan <- true
				log.ready_file_chan <- true
			}
			for _, ctrl := range(sheep.controls) {
				ctrl.ready_request_chan <- true
				ctrl.init(ctrl.knob, ctrl_get_remote, ctrl_set_remote)
			}
		}
		bayopt_m.show()
	}
}


/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of a Baysian optimization shepherd.
   - bayopt_shepherd expects logs to consist of 99th tail latency + total joules consumed 
     to represent execution for some period of time
*/
func (bayopt_s bayopt_shepherd) process_logs(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				l_m := l_m
				sheep := sheep
				log := log
				fmt.Printf("\033[32m-------- PROCESS LOG SIGNAL :  %v - %v - %v\n\033[0m", m_id, sheep_id, log_id)
				mem_buff := *(log.mem_buff)

				// persist log in local log file
				<- log.ready_file_chan
				sheep.update_log_file(log.id)
				select { 
				case log.ready_file_chan <- true:
				default:
				}

				joules_val := float32(mem_buff[0][joules_idx]) * 0.000061
				sheep.perf_data["joules_measure"] = append(sheep.perf_data["joules_measure"], joules_val)
				joules_old := sheep.perf_data["joules_measure"][len(sheep.perf_data["joules_measure"]) - 2]
				sheep.perf_data["joules_diff"] = append(sheep.perf_data["joules_diff"], joules_val - joules_old)

				timestamp_val := float32(mem_buff[0][timestamp_idx]) * 1./(2899999*1000)
				sheep.perf_data["timestamp_measure"] = append(sheep.perf_data["timestamp_measure"], timestamp_val)

				fmt.Printf("\033[35m---------- TIMESTAMP MEAS: %v\n\033[0m", sheep.perf_data["timestamp_measure"])
				fmt.Printf("\033[33m---------- JOULES MEAS: %v\n\033[0m", sheep.perf_data["joules_measure"])
				fmt.Printf("\033[33m---------- JOULES DIFF: %v\n\033[0m", sheep.perf_data["joules_diff"])

				// muster can now overwrite mem_buff for this log
				select {
				case sheep.logs[log.id].ready_process_chan <- true:
				default:
				}

				//if len(bayopt_s.joules_diff[l_m.id][sheep.id]) % 2 == 0 {
				if len(sheep.perf_data["joules_diff"]) % 2 == 0 {
					sheep.ready_ctrl_chan <- true
				}

				fmt.Printf("\033[32m-------- COMPLETED PROCESS LOG :  %v - %v - %v\n\033[0m", l_m.id, sheep.id, log.id)	
			} ()
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

/* 
  This function implements the control computation loop of a Bayesian optimization shepherd.
*/
func (bayopt_s bayopt_shepherd) compute_control(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	var ctrl_dvfs_id string
	var ctrl_itr_id string
	var ctrl_dvfs_val uint64
	var ctrl_itr_val uint64
	for {
		select {
		case opt_req := <- l_m.request_optimize_chan:
			fmt.Printf("\033[31m-------- REQUEST OPTIMIZE SIGNAL :  %v - %v\n\033[0m", m_id, opt_req)
			ctrls := opt_req.settings
			for _, ctrl := range(ctrls) {
				switch {
				case ctrl.knob == "dvfs":
					ctrl_dvfs_id = "dvfs-ctrl-"
					ctrl_dvfs_val = ctrl.val
				case ctrl.knob == "itr-delay":
					ctrl_itr_id = "itr-ctrl-" + l_m.ip
					ctrl_itr_val = ctrl.val
				default:
					fmt.Println("****** Unimplemented optimization control setting: ", ctrl)
				}
			}

			// set settings at remote muster
			for _, sheep := range(l_m.pasture) {
				c_str := strconv.Itoa(int(sheep.core))
				start_ctrls := make(map[string]uint64)
				start_ctrls[ctrl_itr_id] = ctrl_itr_val
				ctrl_dvfs_id = "dvfs-ctrl-" + c_str + "-" + l_m.ip
				start_ctrls[ctrl_dvfs_id] = ctrl_dvfs_val
				bayopt_s.control(l_m.id, sheep.id, start_ctrls)
			}

			// at this point, ctrl values are set in local muster representation
			bayopt_s.init_log_files(bayopt_s.bayopt_musters[m_id].logs_dir)
			l_m.ready_optimize_chan <- true
		}
	}
}






