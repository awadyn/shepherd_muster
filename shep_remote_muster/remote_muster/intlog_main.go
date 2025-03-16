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
	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.logs_dir = home_dir + "/" + r_m.id + ".intlog_logs/"
	_, err = os.Stat(intlog_m.logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(intlog_m.logs_dir, 0777)
		if err != nil { panic(err) }
	}
	intlog_m.init()
	intlog_m.init_remote()

	intlog_m.show()
	intlog_m.deploy()

//	for sheep_id, _ := range(intlog_m.pasture) {
//		go intlog_m.ctrl_manage(sheep_id) 
//	}

	// cleanup
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}



