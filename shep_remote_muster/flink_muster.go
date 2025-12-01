package main

import (
	"strconv"
)

/*********************************************/

func (source_m *flink_source_muster) init() {
	source_m.flink_metrics = []string{"i", "backpressure", "timestamp"}
	source_m.buff_max_size = 1
	for sheep_id, sheep := range(source_m.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "backpressure-log-" + c_str + "-" + source_m.ip
		log_c := log{id: log_id}
		log_c.init(source_m.buff_max_size, source_m.flink_metrics, 3)
		source_m.pasture[sheep_id].logs[log_id] = &log_c
	}
	source_m.init_log_files(source_m.logs_dir)
}


