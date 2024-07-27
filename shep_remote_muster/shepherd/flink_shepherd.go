package main

import (
//	"fmt"
	"strconv"
//	"math/rand"
//	"slices"
)

/**************************************/

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/


/* 
   This function initializes a specialized shepherd for energy-and-performance 
   supervision. Each muster under this shepherd's supervision logs energy and
   performance metrics - i.e. joules and timestamp counters - and controls
   energy and performance settings - i.e. interrupt delay 
   and dynamic-voltage-frequency-scaling - for each core under its supervision.
*/
func (flink_s *flink_shepherd) init() {
	flink_s.logs_dir = "/users/awadyn/shepherd_muster/flink_logs/"
	flink_s.flink_metrics = []string{"joules", "backpressure"}
	flink_s.buff_max_size = 1
	for m_id, m := range(flink_s.musters) {
		var c uint8
		for c = 0; c < m.ncores; c++ {
			c_str := strconv.Itoa(int(c))
			sheep_id := c_str + "-" + m.ip
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip,
					metrics: flink_s.flink_metrics, 
					max_size: flink_s.buff_max_size, 
					ready_request_chan: make(chan bool, 1),
					ready_buff_chan: make(chan bool, 1),
					ready_process_chan: make(chan bool, 1)}
			mem_buff := make([][]uint64, 0)
			log_c.mem_buff = &mem_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, knob: "dvfs", value: 0xc00, dirty: false}
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, knob: "itr-delay", value: 1, dirty: false}

			flink_s.musters[m_id].pasture[sheep_id].logs[log_id] = &log_c
			flink_s.musters[m_id].pasture[sheep_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			flink_s.musters[m_id].pasture[sheep_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
			flink_s.musters[m_id].pasture[sheep_id].logs[log_c.id].ready_request_chan <- true
			flink_s.musters[m_id].pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
		}
	}
	flink_s.init_log_files(flink_s.logs_dir)
}


