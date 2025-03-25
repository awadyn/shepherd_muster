package main

import (
//	"os"
	"fmt"
//	"slices"
//	"sort"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

type bayopt_intlog_shepherd struct {
	shepherd
	bayopt_intlog_musters map[string]*bayopt_intlog_muster 
}


func (bayopt_s *bayopt_intlog_shepherd) init() {
	bayopt_s.bayopt_intlog_musters = make(map[string]*bayopt_intlog_muster)
	for _, l_m := range(bayopt_s.local_musters) {
		intlog_m := intlog_muster{local_muster: *l_m}
		intlog_m.init()

		bayopt_m := bayopt_intlog_muster{intlog_muster: intlog_m}
		bayopt_m.init()

		bayopt_s.bayopt_intlog_musters[bayopt_m.id] = &bayopt_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (bayopt_s bayopt_intlog_shepherd) process_logs(m_id string) {
	l_m := bayopt_s.bayopt_intlog_musters[m_id]
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
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id) }
				select {
				case log.ready_process_chan <- true:
				default:
				}
				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }

			} ()
		}
	}
}







