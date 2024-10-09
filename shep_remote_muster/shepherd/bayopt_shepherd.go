package main

import (
	"fmt"
	"strconv"
	"math/rand"
	"slices"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

var joules_idx int
var timestamp_idx int

type bayopt_muster struct {
	local_muster
	logs_dir string
	ixgbe_metrics []string
	buff_max_size uint64
}

type bayopt_shepherd struct {
	shepherd
	bayopt_musters map[string]*bayopt_muster 
	logs_dir string
	ixgbe_metrics []string
	buff_max_size uint64
	// joules_measure[muster_id][sheep_id] = X
	joules_measure map[string](map[string][]float64)
	// joules_diff[muster_id][sheep_id] = Y
	joules_diff map[string](map[string][]float64)
}

/* 
   This function initializes a specialized shepherd for energy-and-performance 
   management. Each muster under shepherd supervision produces energy and
   performance logs and applies control changes to energy and performance 
   settings - i.e. interrupt delay and dynamic-voltage-frequency-scaling - 
   upon shepherd decision-making for each core under muster supervision.
*/
func (bayopt_s *bayopt_shepherd) init() {
	bayopt_s.ixgbe_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
	                                "instructions", "cycles", "ref_cycles", "llc_miss",
	                                "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	bayopt_s.buff_max_size = 1
	bayopt_s.bayopt_musters = make(map[string]*bayopt_muster)
	joules_idx = slices.Index(bayopt_s.ixgbe_metrics, "joules")
	timestamp_idx = slices.Index(bayopt_s.ixgbe_metrics, "timestamp")
	bayopt_s.joules_measure = make(map[string](map[string][]float64))
	bayopt_s.joules_diff = make(map[string](map[string][]float64))

	for _, l_m := range(bayopt_s.local_musters) {
		bayopt_s.joules_measure[l_m.id] = make(map[string][]float64)
		bayopt_s.joules_diff[l_m.id] = make(map[string][]float64)
		bayopt_m := bayopt_muster{local_muster: *l_m}
		bayopt_m.init()
		bayopt_s.bayopt_musters[bayopt_m.id] = &bayopt_m
	}
}

func (bayopt_s *bayopt_shepherd) init_local() {
	for _, bayopt_m := range(bayopt_s.bayopt_musters) {
		for _, sheep := range(bayopt_m.pasture) {
			bayopt_s.joules_measure[bayopt_m.id][sheep.id] = make([]float64, 1)
			bayopt_s.joules_diff[bayopt_m.id][sheep.id] = make([]float64, 0)
			for _, log := range(sheep.logs) {
				log.ready_process_chan <- true
				log.ready_request_chan <- true
				log.ready_buff_chan <- true
			}
		}
		bayopt_m.show()
	}
}


/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (sheep_c *sheep) update_log_file(log_id string) {
	str_mem_buff := make([][]string,0)
	log := sheep_c.logs[log_id]
	mem_buff := *(log.mem_buff)
	for _, row := range(mem_buff) {
		if len(row) == 0 { break }
		str_row := []string{}
		for i := range(len(log.metrics)) {
			val := strconv.Itoa(int(row[i]))
			str_row = append(str_row, val)
		}
		str_mem_buff = append(str_mem_buff, str_row)
	}
	writer := sheep_c.log_writer_map[log_id]
	writer.WriteAll(str_mem_buff)
	fmt.Println(mem_buff)
	fmt.Println(str_mem_buff)
}

/* 
   This function implements the log processing loop of a Baysian optimization shepherd.
   - bayopt_shepherd expects logs to consist of 99th tail latency + total joules consumed 
     to represent execution for some period of time
*/
func (bayopt_s bayopt_shepherd) process_logs() {
	for {
		select {
		case ids := <- bayopt_s.process_buff_chan:
			m_id := ids[0]
			sheep_id := ids[1]
			log_id := ids[2]
			l_m := bayopt_s.local_musters[m_id]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				l_m := l_m
				sheep := sheep
				log := log
				fmt.Printf("\033[32m-------- PROCESS LOG SIGNAL :  %v - %v - %v\n\033[0m", m_id, sheep_id, log_id)
				mem_buff := *(log.mem_buff)
				sheep.update_log_file(log.id)
				joules_val := float64(mem_buff[0][joules_idx]) * 0.000061
				bayopt_s.joules_measure[m_id][sheep_id] = append(bayopt_s.joules_measure[m_id][sheep_id], joules_val)
				joules_old := bayopt_s.joules_measure[m_id][sheep_id][len(bayopt_s.joules_measure[m_id][sheep_id]) - 2]
				bayopt_s.joules_diff[m_id][sheep_id] = append(bayopt_s.joules_diff[m_id][sheep_id], joules_val - joules_old)
				fmt.Printf("\033[33m---------- %v\n\033[0m", mem_buff)
				fmt.Printf("\033[33m---------- JOULES MEAS: %v\n\033[0m", bayopt_s.joules_measure[l_m.id][sheep.id])
				fmt.Printf("\033[33m---------- JOULES DIFF: %v\n\033[0m", bayopt_s.joules_diff[l_m.id][sheep.id])
				fmt.Printf("\033[32m-------- COMPLETED PROCESS LOG :  %v - %v - %v\n\033[0m", l_m.id, sheep.id, log.id)	
				// muster can now overwrite mem_buff for this log
				select {
				case sheep.logs[log.id].ready_process_chan <- true:
				default:
				}
//				if len(bayopt_s.joules_diff[l_m.id][sheep.id]) % 2 == 0 {
//					select {
//					case bayopt_s.compute_ctrl_chan <- []string{l_m.id, sheep.id}:
//					default:
//					}
//				}
			} ()
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

func (bayopt_s *bayopt_shepherd) bayopt_ctrl(m_id string, sheep_id string) map[string]uint64 {
	new_ctrls := make(map[string]uint64)
	m := bayopt_s.musters[m_id]
	sheep := m.pasture[sheep_id]
	c_str := strconv.Itoa(int(sheep.core))
	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip

	dvfs_list := []uint64{0xc00, 0xe00, 0x1100, 0x1300, 0x1500, 0x1700, 0x1900, 0x1a00}
//	itr_list := []uint64{2, 100, 400}
	dvfs_idx := rand.Intn(len(dvfs_list))
	new_dvfs := dvfs_list[dvfs_idx]
//	itr_idx := rand.Intn(len(itr_list))
//	new_itr := itr_list[itr_idx]

	new_ctrls[ctrl_dvfs_id] = new_dvfs
//	new_ctrls[ctrl_itr_id] = new_itr
	return new_ctrls
}

/* 
  This function implements the control computation loop of a Bayesian optimization shepherd.
*/
func (bayopt_s bayopt_shepherd) compute_control() {
	for {
		select {
		case ids := <- bayopt_s.compute_ctrl_chan:
			m_id := ids[0]
			sheep_id := ids[1]
			l_m := bayopt_s.local_musters[m_id]
			sheep := l_m.pasture[sheep_id]
			go func() {
				l_m := l_m
				sheep := sheep
				new_ctrls := bayopt_s.bayopt_ctrl(l_m.id, sheep.id)
				fmt.Printf("\033[35m<------- CTRL REQ --  %v - %v - %v\n\033[0m", l_m.id, sheep.id, new_ctrls)
				l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: new_ctrls}
				ctrl_reply := <- sheep.ready_ctrl_chan
				ctrls := ctrl_reply.ctrls
				done_ctrl := ctrl_reply.done
				if done_ctrl { 
					for ctrl_id, ctrl_val := range(ctrls) {
						sheep.controls[ctrl_id].value = ctrl_val
					}
				}
			} ()
		}
	}
}






