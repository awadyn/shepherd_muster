package main

import (
	"fmt"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

type k8s_shepherd struct {
	shepherd
	k8s_musters map[string]*k8s_muster
}


func (k8s_s *k8s_shepherd) init() {
	k8s_s.k8s_musters = make(map[string]*k8s_muster)

	for _, l_m := range(k8s_s.local_musters) {
		k8s_m := k8s_muster{local_muster: *l_m}
		k8s_m.init()
		k8s_s.k8s_musters[l_m.id] = &k8s_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/


func (k8s_s k8s_shepherd) process_logs(m_id string) {
	l_m := k8s_s.k8s_musters[m_id]
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
			} ()
		}
	}
}


func (k8s_s k8s_shepherd) process_control() {
	for {
		select {
		case <- k8s_s.new_ctrl_chan:
			if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS CONTROL SIGNAL :  \n\033[0m") }
		}
	}
}




