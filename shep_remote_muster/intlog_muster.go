package main

import (
	"strconv"
)
/*********************************************/


func (intlog_m *intlog_muster) init() {
	intlog_m.ixgbe_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	intlog_m.buff_max_size = 4096

	ctrl_itr_shared := control{id: "itr-ctrl-" + intlog_m.ip, n_ip: intlog_m.ip}
//	ctrl_itr_shared.init("itr-delay", read_rx_usecs, write_rx_usecs)
	for sheep_id, sheep := range(intlog_m.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "ixgbe-log-" + c_str + "-" + intlog_m.ip
		log_c := log{id: log_id}
		log_c.init(intlog_m.buff_max_size, intlog_m.ixgbe_metrics, 2)
		ctrl_dvfs := control{id: "dvfs-ctrl-" + c_str + "-" + intlog_m.ip, n_ip: intlog_m.ip}
		ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
		var ctrl_itr control
		ctrl_itr = ctrl_itr_shared
		ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
		intlog_m.pasture[sheep_id].logs[log_id] = &log_c	
		intlog_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		intlog_m.pasture[sheep_id].controls[ctrl_itr_shared.id] = &ctrl_itr
	}
}

