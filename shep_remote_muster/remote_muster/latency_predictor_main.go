package main

import (
	"os"
	"time"
)

/*********************************************/

func latency_predictor_main(n node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m := muster{node: n}
	m.init()

	m.logs_dir = home_dir + "/" + m.id + ".intlog_logs/"
	err = os.Mkdir(m.logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }

	r_m := remote_muster{muster: m}
	r_m.init()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()
	intlog_m.init_remote()

	latency_predictor_m := latency_predictor_muster{intlog_muster: intlog_m}
	latency_predictor_m.init()
	latency_predictor_m.show()

	go latency_predictor_m.start_pulser()
	go latency_predictor_m.start_controller()
	go latency_predictor_m.start_coordinator()
	latency_predictor_m.start_logger()

	for sheep_id, _ := range(latency_predictor_m.pasture) {
		go latency_predictor_m.ctrl_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	latency_predictor_m.cleanup()
}

