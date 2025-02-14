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
		intlog_m := intlog_muster{local_muster: *l_m}
		intlog_m.init()
		intlog_s.intlog_musters[intlog_m.id] = &intlog_m

		intlog_m.logs_dir = logs_dir + intlog_m.id + "/"
		err := os.Mkdir(intlog_m.logs_dir, 0750)
		if err != nil && !os.IsExist(err) { panic(err) }
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of a Baysian optimization shepherd.
   - bayopt_shepherd expects logs to consist of 99th tail latency + total joules consumed 
     to represent execution for some period of time
*/
func (intlog_s intlog_shepherd) process_logs(m_id string) {
	l_m := intlog_s.local_musters[m_id]
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
				fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id)
				select {
				case sheep.logs[log.id].ready_process_chan <- true:
				default:
				}

				fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id)	
			} ()
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

//func (bayopt_s *bayopt_shepherd) bayopt_ctrl(m_id string, sheep_id string) map[string]uint64 {
//	new_ctrls := make(map[string]uint64)
//	m := bayopt_s.musters[m_id]
//	sheep := m.pasture[sheep_id]
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
////	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
//
//	dvfs_list := []uint64{0xc00, 0xe00, 0x1100, 0x1300, 0x1500, 0x1700, 0x1900, 0x1a00}
////	itr_list := []uint64{2, 100, 400}
//	dvfs_idx := rand.Intn(len(dvfs_list))
//	new_dvfs := dvfs_list[dvfs_idx]
////	itr_idx := rand.Intn(len(itr_list))
////	new_itr := itr_list[itr_idx]
//
//	new_ctrls[ctrl_dvfs_id] = new_dvfs
////	new_ctrls[ctrl_itr_id] = new_itr
//	return new_ctrls
//}
//
///* 
//  This function implements the control computation loop of a Bayesian optimization shepherd.
//*/
func (intlog_s intlog_shepherd) compute_control(m_id string) {
//	for {
//		select {
//		case ids := <- bayopt_s.compute_ctrl_chan:
//			m_id := ids[0]
//			sheep_id := ids[1]
//			l_m := bayopt_s.local_musters[m_id]
//			sheep := l_m.pasture[sheep_id]
//			go func() {
//				l_m := l_m
//				sheep := sheep
//				new_ctrls := bayopt_s.bayopt_ctrl(l_m.id, sheep.id)
//				fmt.Printf("\033[35m<------- CTRL REQ --  %v - %v - %v\n\033[0m", l_m.id, sheep.id, new_ctrls)
//				l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: new_ctrls}
//				ctrl_reply := <- sheep.ready_ctrl_chan
//				ctrls := ctrl_reply.ctrls
//				done_ctrl := ctrl_reply.done
//				if done_ctrl { 
//					for ctrl_id, ctrl_val := range(ctrls) {
//						sheep.controls[ctrl_id].value = ctrl_val
//					}
//				}
//			} ()
//		}
//	}
}






