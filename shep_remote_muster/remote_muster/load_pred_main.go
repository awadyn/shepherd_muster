package main

import (
//	"os"
	"time"
)

/*********************************************/

func load_pred_main(n node) {
	m := muster{node: n}
	m.init()
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()
	
	load_pred_m := load_pred_muster{intlog_muster: intlog_m}
	load_pred_m.init()

	load_pred_m.init_log_files(intlog_m.logs_dir)
	load_pred_m.show()
	load_pred_m.deploy()

	for sheep_id, _ := range(load_pred_m.pasture) {
		go load_pred_m.ctrl_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	load_pred_m.cleanup()
}



