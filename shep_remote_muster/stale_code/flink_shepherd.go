package main

import (
	"fmt"
	"strconv"
//	"math/rand"
//	"slices"
)

/**************************************/

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

func (flink_s *flink_shepherd) init() {
	flink_s.flink_musters = make(map[string]*flink_local_muster)
	for _, l_m := range(flink_s.local_musters) {
		flink_m := &flink_local_muster{base_muster: l_m, 
					       logs_dir: "/users/awadyn/shepherd_muster/shepherd_flink_logs/"}
		flink_s.flink_musters[flink_m.base_muster.id] = flink_m
		switch {
		case flink_m.base_muster.ip_idx == 0:
			flink_m.base_muster.role = "worker"
			flink_s.init_worker_muster(flink_m.base_muster.id)
		case flink_m.base_muster.ip_idx == 1:
			flink_m.base_muster.role = "source"
			flink_s.init_source_muster(flink_m.base_muster.id)
		default:
		}
		flink_m.base_muster.show()	
	}
}

func (flink_s *flink_shepherd) init_worker_muster(m_id string) {
	flink_m := flink_s.flink_musters[m_id]
	flink_m.flink_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
	                                 "instructions", "cycles", "ref_cycles", "llc_miss",
	                                 "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	flink_m.buff_max_size = 1
	for sheep_id, sheep := range(flink_m.base_muster.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "log-" + c_str + "-" + flink_m.base_muster.ip 
		log_c := log{id: log_id, n_ip: flink_m.base_muster.ip,
			     metrics: flink_m.flink_metrics, 
			     max_size: flink_m.buff_max_size, 
			     ready_request_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1),
			     ready_process_chan: make(chan bool, 1)}
		mem_buff := make([][]uint64, 0)
		log_c.mem_buff = &mem_buff
		ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + flink_m.base_muster.ip
		ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: flink_m.base_muster.ip, knob: "dvfs", dirty: false, ready_request_chan: make(chan bool, 1)}
		ctrl_itr_id := "ctrl-itr-" + c_str + "-" + flink_m.base_muster.ip
		ctrl_itr_c := control{id: ctrl_itr_id, n_ip: flink_m.base_muster.ip, knob: "itr-delay", dirty: false, ready_request_chan: make(chan bool, 1)}		
		flink_m.base_muster.pasture[sheep_id].logs[log_id] = &log_c
		flink_m.base_muster.pasture[sheep_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
		flink_m.base_muster.pasture[sheep_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
		flink_m.base_muster.pasture[sheep_id].logs[log_c.id].ready_request_chan <- true
		flink_m.base_muster.pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
	}
}

func (flink_s *flink_shepherd) init_source_muster(m_id string) {
	flink_m := flink_s.flink_musters[m_id]
	flink_m.flink_metrics = []string{"i", "backpressure","timestamp"}
	flink_m.buff_max_size = 1
	for sheep_id, sheep := range(flink_m.base_muster.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "log-" + c_str + "-" + flink_m.base_muster.ip 
		log_c := log{id: log_id, n_ip: flink_m.base_muster.ip,
			     metrics: flink_m.flink_metrics, 
			     max_size: flink_m.buff_max_size, 
			     ready_request_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1),
			     ready_process_chan: make(chan bool, 1)}
		mem_buff := make([][]uint64, 0)
		log_c.mem_buff = &mem_buff
		flink_m.base_muster.pasture[sheep_id].logs[log_id] = &log_c
		flink_m.base_muster.pasture[sheep_id].logs[log_c.id].ready_request_chan <- true
		flink_m.base_muster.pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
	}
}


/* This function starts all muster threads required by a shepherd. */
func (flink_s *flink_shepherd) deploy_musters() {
	for _, l_m := range(flink_s.local_musters) {
		fmt.Printf("\033[97;1m**** DEPLOYING MUSTER %v ****\n\033[0m", l_m.id)
		l_m.start_pulser()					// per-muster pulse client
		if l_m.role == "worker" { l_m.start_controller() }	// per-muster ctrl client
		l_m.start_coordinator()					// per-muster coordinate client
		l_m.start_logger()					// per-muster log server
	}
}



