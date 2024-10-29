package main

import (
	"strconv"
)

/*********************************************/


func (bayopt_m *bayopt_muster) init() {
	bayopt_m.ixgbe_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	bayopt_m.buff_max_size = 1

	ctrl_itr_shared := control{id: "itr-ctrl-" + bayopt_m.ip, n_ip: bayopt_m.ip}
	for sheep_id, sheep := range(bayopt_m.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "ixgbe-log-" + c_str + "-" + bayopt_m.ip
		log_c := log{id: log_id}
		log_c.init(bayopt_m.buff_max_size, bayopt_m.ixgbe_metrics, 3)
		ctrl_dvfs := control{id: "dvfs-ctrl-" + c_str + "-" + bayopt_m.ip, n_ip: bayopt_m.ip}
		ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
		var ctrl_itr control
		ctrl_itr = ctrl_itr_shared
		ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
		bayopt_m.pasture[sheep_id].logs[log_id] = &log_c	
		bayopt_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		bayopt_m.pasture[sheep_id].controls[ctrl_itr_shared.id] = &ctrl_itr
	}
}

