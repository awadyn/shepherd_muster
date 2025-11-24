package main

import (
	"strconv"
	"time"
)
/*********************************************/

//var ixgbe_metrics []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
//			    "instructions", "cycles", "ref_cycles", "llc_miss", 
//			    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var ixgbe_metrics []string = []string{"rx_bytes", "timestamp"}
var buff_max_size uint64 = 1024
var log_wait_factor time.Duration = 1

func (intlog_m *intlog_muster) init() {
	intlog_m.native_loggers["intlogger"] = ixgbe_native_log
	for _, sheep := range(intlog_m.pasture) {
		// specialize default log id then init log specifics
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label

		log_id := "log-" + label + "-" + index
		log_c := sheep.logs[log_id]
		new_log_id := "ixgbe-log-" + label + "-" + index + "-" + intlog_m.ip
		log_c.id = new_log_id
		log_c.init_specs(buff_max_size, ixgbe_metrics, log_wait_factor)

		sheep.logs[new_log_id] = log_c
		delete(sheep.logs, log_id)
	}
}

