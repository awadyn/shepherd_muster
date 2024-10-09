package main

import (
	"os"
	"time"
)

/*********************************************/

func bayopt_main(n node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m := muster{node: n}
	m.init()
	m.init_remote()
	r_m := remote_muster{muster: m}
	r_m.init()

	bayopt_m := bayopt_muster{remote_muster: r_m, 
				  logs_dir: home_dir + "/" + r_m.id + ".bayopt_logs/"}
	_, err = os.Stat(bayopt_m.logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(bayopt_m.logs_dir, 0777)
		if err != nil { panic(err) }
	}
	bayopt_m.init()
	bayopt_m.init_remote()
	bayopt_m.show()

	go bayopt_m.start_pulser()
//	go bayopt_m.start_controller()
//	go bayopt_m.sync_new_ctrl()
	go bayopt_m.start_coordinator()
	bayopt_m.start_logger()

	for sheep_id, _ := range(bayopt_m.pasture) {
		go bayopt_m.log_manage(sheep_id, bayopt_m.logs_dir, ixgbe_native_log) 
//		go bayopt_m.ctrl_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	bayopt_m.cleanup()
}



