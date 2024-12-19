package main

import (
	"fmt"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

type nop_muster struct {
	local_muster
	logs_dir string
}

type nop_shepherd struct {
	shepherd
	nop_musters map[string]*nop_muster 
	logs_dir string
}

func (nop_s *nop_shepherd) init() {
	nop_s.nop_musters = make(map[string]*nop_muster)

	for _, l_m := range(nop_s.local_musters) {
		nop_m := nop_muster{local_muster: *l_m}
		nop_m.init()
		nop_s.nop_musters[nop_m.id] = &nop_m
	}
}

func (nop_s *nop_shepherd) init_local() {
	for _, nop_m := range(nop_s.nop_musters) {
		for _, sheep := range(nop_m.pasture) {
			for _, log := range(sheep.logs) {
				log.ready_process_chan <- true
				log.ready_request_chan <- true
				log.ready_buff_chan <- true
			}
		}
		nop_m.show()
	}
}


/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (nop_s nop_shepherd) process_logs(m_id string) {
	l_m := nop_s.local_musters[m_id]
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				l_m := l_m
				sheep := sheep
				log := log
				fmt.Printf("\033[32m-------- PROCESS LOG SIGNAL :  %v - %v - %v\n\033[0m", m_id, sheep_id, log_id)
				fmt.Printf("\033[32m-------- COMPLETED PROCESS LOG :  %v - %v - %v\n\033[0m", l_m.id, sheep.id, log.id)	
				// muster can now overwrite mem_buff for this log
				select {
				case sheep.logs[log.id].ready_process_chan <- true:
				default:
				}
			} ()
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

func (nop_s nop_shepherd) compute_control(m_id string) {
	l_m := nop_s.local_musters[m_id]
	for {
		select {
		case opt_req := <- l_m.request_optimize_chan:
			fmt.Printf("\033[31m-------- REQUEST OPTIMIZE SIGNAL :  %v - %v\n\033[0m", m_id, opt_req)
		}
	}
}






