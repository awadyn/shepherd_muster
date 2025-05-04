package main

import (
//	"os"
	"time"
)

/*********************************************/

func intlog_main(n node) {
	m := muster{node: n}
	m.init()
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()
	intlog_m.init_log_files(intlog_m.logs_dir)
	intlog_m.show()
	intlog_m.deploy()

//	for sheep_id, _ := range(intlog_m.pasture) {
//		go intlog_m.ctrl_manage(sheep_id) 
//	}

	// cleanup
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}



