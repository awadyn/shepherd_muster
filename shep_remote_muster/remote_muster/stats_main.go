package main

import (
//	"os"
	"time"
)

/*********************************************/

// a stats_muster is an intlog_muster with additional data structures
// to store useful log statistics
type stats_muster struct {
	intlog_muster

	rx_bytes_all map[string][]uint64
	timestamps_all map[string][]uint64
	rx_bytes_concat []int
	rx_bytes_medians []int
	rx_bytes_median int
	processing_lock chan bool
}

func stats_main(n node) {
	m := muster{node: n}
	m.init()
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()
	intlog_m.init_log_files(intlog_m.logs_dir)

	stats_m := stats_muster{intlog_muster: intlog_m}
	stats_m.init()
	stats_m.init_remote()
	stats_m.show()
	stats_m.deploy()

	for sheep_id, sheep := range(stats_m.pasture) {
		if sheep.label == "node" { go stats_m.ctrl_manage(sheep_id) }
	}

	// cleanup
	time.Sleep(exp_timeout)
	stats_m.cleanup()
}


