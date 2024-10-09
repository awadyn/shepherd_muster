package main

import (
	"os"
	"time"
)

/*********************************************/

func nop_main(n node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m := muster{node: n}
	m.init()
	m.init_remote()
	r_m := remote_muster{muster: m}
	r_m.init()

	nop_m := nop_muster{remote_muster: r_m, 
			    logs_dir: home_dir + "/" + r_m.id + ".nop_logs/"}
	_, err = os.Stat(nop_m.logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(nop_m.logs_dir, 0777)
		if err != nil { panic(err) }
	}
	nop_m.init()
	nop_m.init_remote()
	nop_m.show()

	go nop_m.start_pulser()
	go nop_m.start_controller()
	go nop_m.sync_new_ctrl()
	go nop_m.start_coordinator()
	nop_m.start_logger()

	for sheep_id, _ := range(nop_m.pasture) {
		go nop_m.log_manage(sheep_id, nop_m.logs_dir, nop_native_log) 
		go nop_m.ctrl_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	nop_m.cleanup()
}



