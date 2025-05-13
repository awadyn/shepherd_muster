package main

import (
//	"os"
	"time"
)

/*********************************************/

func bayopt_intlog_main(n node) {
	m := muster{node: n}
	m.init()
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()

	bayopt_m := bayopt_intlog_muster{intlog_muster: intlog_m}
	bayopt_m.init()
	bayopt_m.init_remote()
	bayopt_m.init_log_files(bayopt_m.logs_dir)
	bayopt_m.show()

	go bayopt_m.start_pulser()
	go bayopt_m.start_controller()
	go bayopt_m.start_coordinator()
	bayopt_m.start_logger()

//	for sheep_id, _ := range(bayopt_m.pasture) {
//		go bayopt_m.ctrl_manage(sheep_id) 
//	}

	// cleanup
	time.Sleep(exp_timeout)
	bayopt_m.cleanup()
}



