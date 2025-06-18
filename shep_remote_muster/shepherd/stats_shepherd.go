package main

import (
//	"os"
	"fmt"
	"slices"
	"sort"
	"os/exec"
	"math"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

// a stats_muster is an intlog_muster with additional data structures
// to store useful log statistics
type stats_muster struct {
	intlog_muster

	rx_bytes_all map[string][]uint64
	timestamps_all map[string][]uint64
	rx_bytes_concat []int
	rx_bytes_medians []int
	rx_bytes_median int
	processing_lock chan bool
}

type stats_shepherd struct {
	intlog_shepherd
	stats_musters map[string]*stats_muster
	rx_bytes_medians map[int][]int
}


func (stats_s *stats_shepherd) init() {
	stats_s.stats_musters = make(map[string]*stats_muster)
	stats_s.rx_bytes_medians = make(map[int][]int)

	for _, intlog_m := range(stats_s.intlog_musters) {
		stats_m := stats_muster{intlog_muster: *intlog_m}
		stats_m.init()
		stats_s.stats_musters[stats_m.id] = &stats_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of a stats shepherd.
   Currently, this shepherd is specialized to 1) compute a metric's signal that
   concatenates this metric's signals from all request servicing cores (e.g. rx_bytes signal) 
   and 2) identify the signal by its median value.

   TODO ..
*/
func (stats_s stats_shepherd) process_logs(m_id string) {
	l_m := stats_s.stats_musters[m_id]
	ctrl_break := 1
	cur_qps_guess := -1
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

				rx_bytes, timestamps := stats_s.get_rx_signal(log)
				
				<- l_m.processing_lock
				// append per-sheep rx_bytes signal to stats_muster data map
				l_m.timestamps_all[sheep_id] = append(l_m.timestamps_all[sheep_id], timestamps...)
				l_m.rx_bytes_all[sheep_id] = append(l_m.rx_bytes_all[sheep_id], rx_bytes...)

				// check if rx_bytes_all complete and concat
				ready := true
				if len(l_m.rx_bytes_all) == len(l_m.pasture) {
					for sheep_id, _ := range(l_m.pasture) {
						if len(l_m.rx_bytes_all[sheep_id]) < 1024 { ready = false }
					}
					if ready {
						// concat
						iterators := make(map[string]int)
						for sheep_id, _ := range(l_m.pasture) { 
							iterators[sheep_id] = 0 
						}
	
						// choose longest sheep log as reference log
						max_size := 0
						min_size := 5000
						ref_sheep := ""
						for sheep_id, _ := range(l_m.pasture) {
							if len(l_m.rx_bytes_all[sheep_id]) > max_size { 
								max_size = len(l_m.rx_bytes_all[sheep_id]) 
								ref_sheep = sheep_id
							}
							if len(l_m.rx_bytes_all[sheep_id]) < min_size { 
								min_size = len(l_m.rx_bytes_all[sheep_id])
							}
						}
	
						l_m.rx_bytes_concat = make([]int, len(l_m.rx_bytes_all[ref_sheep]))
						for j := 0; j < max_size; j ++ {
							for sheep_id, _ := range(l_m.pasture) {
								j_itr := iterators[sheep_id]
								i_timestamps := l_m.timestamps_all[sheep_id]
								if j_itr == len(i_timestamps) { continue }
								if i_timestamps[j_itr] <= l_m.timestamps_all[ref_sheep][j] {
									l_m.rx_bytes_concat[j] += int(l_m.rx_bytes_all[sheep_id][j_itr])
									iterators[sheep_id] += 1
								}
							}
						}

						// get current median of the concatenated rx-bytes frame
						sort.Ints(l_m.rx_bytes_concat)
						l_m.rx_bytes_median = l_m.rx_bytes_concat[max_size/2]

						// OFFLINE PHASE: append to medians list of per-frame medians across run
//						l_m.rx_bytes_medians = append(l_m.rx_bytes_medians, l_m.rx_bytes_median)

						/* ONLINE PHASE: testing with manually collected medians */
						var guess int
						var diffs []float64 = []float64{0, 0, 0, 0, 0}
						for q := 0; q < len(qpses); q ++ {
							diffs[q] = math.Abs(float64(l_m.rx_bytes_median - medians[q]))
//							if l_m.rx_bytes_median <= (medians[q] + medians[q]/4) { 
//								guess = q
//								break 
//							}
						}
						var min_diff float64 = 60000
						for q := 0; q < len(qpses); q ++ {
							if diffs[q] < min_diff { 
								min_diff = diffs[q] 
								guess = q
							}
						}

						fmt.Println("************ QPS GUESS -- ", qpses[guess], " -- MEDIAN -- ", l_m.rx_bytes_median)
						if cur_qps_guess == qpses[guess] { 
							ctrl_break = 1 
						} else {
							if ctrl_break == 0 {
								fmt.Println("************ APPLYING CTRLS **********************", opt_dvfs[guess], opt_itrd[guess])
								dvfs_cmd := "sudo wrmsr -a 0x199 " + opt_dvfs[guess]
								itrd_cmd := "sudo ethtool -C enp130s0f0 rx-usecs " + opt_itrd[guess]
								ctrl_cmd := dvfs_cmd + "; " + itrd_cmd	
								cmd := exec.Command("bash", "-c", "ssh -f awadyn@130.127.133.42 '" + ctrl_cmd + "'")
								if err := cmd.Run(); err != nil { panic(err) }
								cur_qps_guess = qpses[guess]
							}
							ctrl_break = (ctrl_break + 1) % 3
						}


//						fmt.Println("ref_sheep: ", ref_sheep, len(l_m.rx_bytes_concat),  "max size:", max_size, "min size:", min_size, "rx_bytes_median: ", l_m.rx_bytes_median)

						// reset/cleanup
						l_m.timestamps_all = make(map[string][]uint64)
						l_m.rx_bytes_all = make(map[string][]uint64)
						for sheep_id, _ := range(l_m.pasture) {
							l_m.timestamps_all[sheep_id] = make([]uint64, 0)
							l_m.rx_bytes_all[sheep_id] = make([]uint64, 0)
						}
						l_m.rx_bytes_concat = make([]int, 0)
					}
				}
				select {
				case l_m.processing_lock <- true:
				default:
				}


				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }

				select {
				case log.ready_process_chan <- true:
				default:
				}
			} ()
		}
	}
}

func (stats_s stats_shepherd) get_rx_signal(l log) ([]uint64, []uint64) {
	// get per-sheep rx_bytes
	mem_buff := l.mem_buff
	var rx_bytes_idx int = slices.Index(l.metrics, "rx_bytes")
	var timestamp_idx int = slices.Index(l.metrics, "timestamp")
	timestamps := make([]uint64, len(*mem_buff))
	rx_bytes := make([]uint64, len(*mem_buff))
	for j := 0; j < len(*mem_buff); j ++ {
		rx_bytes[j] = (*mem_buff)[j][rx_bytes_idx]
		timestamps[j] = (*mem_buff)[j][timestamp_idx]
	}
	return rx_bytes, timestamps
}







