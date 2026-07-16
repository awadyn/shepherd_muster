package main

import (
	"time"
//	"fmt"
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

	/* YA NEW */
//	// TODO remove - temporary test
//	time.Sleep(time.Second * 2)
//	logger_id := "intlogger"
//	for sheep_id, _ := range(intlog_m.pasture) {
//		sheep := r_m.pasture[sheep_id]
//		if sheep.label == "node" {continue}
//		for log_id, _ := range(sheep.logs) {
//			log := sheep.logs[log_id]
//	
//			coordinate_cmd := "start"
//			<- log.ready_request_chan
//			if debug { fmt.Printf("\033[34m---> COORD REQ -- %v -- %v -- %v -- %v\n\033[0m", sheep_id, log_id, coordinate_cmd, logger_id) }
//			cmd_status := r_m.log_manage(sheep.id, log.id, coordinate_cmd, "intlogger")
//			if !cmd_status { fmt.Printf("\033[31;1m****** ERROR: %v failed to send log coordinate request %v for %v\n\033[0m", r_m.id, coordinate_cmd, log.id, logger_id) }
//			if debug { fmt.Printf("\033[34m----DONE COORD REQ %v -- %v -- %v -- %v\n\033[0m", sheep_id, log_id, coordinate_cmd, logger_id) }
//			
//			coordinate_cmd = "all"
//			<- log.ready_request_chan
//			if debug { fmt.Printf("\033[34m---> COORD REQ -- %v -- %v -- %v -- %v\n\033[0m", sheep_id, log_id, coordinate_cmd, logger_id) }
//			cmd_status = r_m.log_manage(sheep.id, log.id, coordinate_cmd, "intlogger")
//			if !cmd_status { fmt.Printf("\033[31;1m****** ERROR: %v failed to send log coordinate request %v for %v\n\033[0m", r_m.id, coordinate_cmd, log.id, logger_id) }
//			if debug { fmt.Printf("\033[34m----DONE COORD REQ %v -- %v -- %v -- %v\n\033[0m", sheep_id, log_id, coordinate_cmd, logger_id) }
//		}
//	}
	/* YA NEW */

	// cleanup
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}



