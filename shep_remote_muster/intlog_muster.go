package main

import (
	"strconv"
	"time"
)
/*********************************************/

var ixgbe_metrics []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
			    "instructions", "cycles", "ref_cycles", "llc_miss", 
			    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var buff_max_size uint64 = 512
var log_wait_factor time.Duration = 2

func (intlog_m *intlog_muster) init() {
	intlog_m.native_loggers["intlogger"] = ixgbe_native_log
	for sheep_id, sheep := range(intlog_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		log_id := "ixgbe-log-" + label + "-" + index + "-" + intlog_m.ip
		log_c := log{id: log_id}
		log_c.init(buff_max_size, ixgbe_metrics, log_wait_factor)
		intlog_m.pasture[sheep_id].logs[log_id] = &log_c	
	}
}

