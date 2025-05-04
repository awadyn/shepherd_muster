package main

import (
	"os"
	"fmt"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

type intlog_muster struct {
	local_muster
}

type intlog_shepherd struct {
	shepherd
	intlog_musters map[string]*intlog_muster 
}

func (intlog_s *intlog_shepherd) init() {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	logs_dir := home_dir + "/" + "intlog-logs-"

	intlog_s.intlog_musters = make(map[string]*intlog_muster)

	for _, l_m := range(intlog_s.local_musters) {
		intlog_m := &intlog_muster{local_muster: *l_m}
		intlog_m.init()
		intlog_s.intlog_musters[intlog_m.id] = intlog_m

		intlog_m.logs_dir = logs_dir + intlog_m.id + "/"

		fmt.Println("local:", l_m.logs_dir)
		fmt.Println("intlog:", intlog_m.logs_dir)

//		intlog_m.logs_dir = logs_dir + intlog_m.id + "/"
		err := os.Mkdir(intlog_m.logs_dir, 0750)
		if err != nil && !os.IsExist(err) { panic(err) }
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of an intlogger shepherd.
   By definition, this shepherd does no log processing further than core log processing.
*/
func (intlog_s intlog_shepherd) process_logs(m_id string) {
	l_m := intlog_s.intlog_musters[m_id]
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				sheep := sheep
				log := log
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep.id, log.id) }
				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }	
				select {
				case log.ready_process_chan <- true:
				default:
				}
			} ()
		}
	}
}






