package main

import (
	"strconv"
)

/*********************************************/


func (flink_m *flink_muster) init() {
	flink_m.logs_dir = "/users/awadyn/shepherd_muster/shep_remote_muster/flink_logs/"
	flink_m.flink_metrics = []string{"joules","backpressure"}
	flink_m.buff_max_size = 1
	var core uint8
	ctrl_itr_shared := control{n_ip: flink_m.ip, knob: "itr-delay", dirty: false}  
	rx_usecs_shared := ctrl_itr_shared.getter(0, read_rx_usecs)  

	for sheep_id, sheep := range(flink_m.pasture) {
		mem_buff := make([][]uint64, flink_m.buff_max_size)
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "log-" + c_str + "-" + flink_m.ip
		log_c := log{id: log_id,
			     metrics: flink_m.flink_metrics,
			     max_size: flink_m.buff_max_size,
			     mem_buff: &mem_buff}
		log_c.init()

		ctrl_dvfs := control{n_ip: flink_m.ip, knob: "dvfs", dirty: false}
		ctrl_dvfs.id = "ctrl-dvfs-" + c_str + "-" + flink_m.ip
		ctrl_dvfs.value = ctrl_dvfs.getter(core, read_dvfs)
		ctrl_itr := ctrl_itr_shared
		ctrl_itr.id = "ctrl-itr-" + c_str + "-" + flink_m.ip
		ctrl_itr.value = rx_usecs_shared

		flink_m.pasture[sheep_id].logs[log_id] = &log_c
		flink_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		flink_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
	}

	flink_m.init_log_files(flink_m.logs_dir)
}



