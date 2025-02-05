package main

import (
	"strconv"
	"time"
)
/*********************************************/
var ixgbe_metrics []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
			    "instructions", "cycles", "ref_cycles", "llc_miss", 
			    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var buff_max_size uint64 = 4096
var log_wait_factor time.Duration = 2

func (intlog_m *intlog_muster) init() {
	for sheep_id, sheep := range(intlog_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		log_id := "ixgbe-log-" + label + "-" + index + "-" + intlog_m.ip
		log_c := log{id: log_id}
		log_c.init(buff_max_size, ixgbe_metrics, log_wait_factor)
		intlog_m.pasture[sheep_id].logs[log_id] = &log_c	
		switch {
		case label == "core":
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + intlog_m.ip, n_ip: intlog_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
			intlog_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + intlog_m.ip, n_ip: intlog_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			intlog_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
		default:
		}
	}
}

