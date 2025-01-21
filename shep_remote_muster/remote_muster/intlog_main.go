package main

import (
	"os"
	"time"
)

/*********************************************/

func intlog_main(n node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m := muster{node: n}
	m.init()
	m.init_remote()
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m,
				  logs_dir: home_dir + "/" + r_m.id + ".intlog_logs/"}
	_, err = os.Stat(intlog_m.logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(intlog_m.logs_dir, 0777)
		if err != nil { panic(err) }
	}
	intlog_m.init()
	intlog_m.init_remote()
	intlog_m.show()

	go intlog_m.start_pulser()
	go intlog_m.start_controller()
	go intlog_m.start_coordinator()
	intlog_m.start_logger()

	for sheep_id, _ := range(intlog_m.pasture) {
		go intlog_m.log_manage(sheep_id, intlog_m.logs_dir, ixgbe_native_log) 
		go intlog_m.ctrl_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}



