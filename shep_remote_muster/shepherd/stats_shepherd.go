package main

import (
//	"os"
	"fmt"
	"slices"
	"sort"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

var medians []int = []int{11035, 19450, 37076, 74224, 112588}
var qpses []int = []int{50000, 100000, 200000, 400000, 600000}

// a stats_muster is an intlog_muster with additional data structures
// to store useful log statistics
type stats_muster struct {
	intlog_muster

	rx_bytes_all map[string][]uint64
	rx_bytes_median int
	processing_lock chan bool
}

type stats_shepherd struct {
	intlog_shepherd
	stats_musters map[string]*stats_muster
	rx_bytes_medians []int
}


func (stats_s *stats_shepherd) init() {
	stats_s.stats_musters = make(map[string]*stats_muster)
	stats_s.rx_bytes_medians = make([]int, 0)

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

				mem_buff := log.mem_buff
				// get per-sheep rx_bytes
				var rx_bytes_idx int = slices.Index(log.metrics, "rx_bytes")
				rx_bytes := make([]uint64, len(*mem_buff))
				for j := 0; j < len(*mem_buff); j ++ {
					rx_bytes[j] = (*mem_buff)[j][rx_bytes_idx]
				}

				<- l_m.processing_lock
				// append per-sheep rx_bytes to muster data 
				l_m.rx_bytes_all[sheep_id] = rx_bytes


				// check if rx_bytes_all complete and compute rx_bytes_total
				if len(l_m.rx_bytes_all) == len(l_m.pasture) - 1 {
//					fmt.Println("COMPUTING TOTAL RX_BYTES SIGNAL..")
					rx_bytes_total := make([]int, 0)
					i := 0
					var total int
					var done_processing int
					for {
						total = 0
						done_processing = 0
						for _, sheep_rx_bytes := range(l_m.rx_bytes_all) {
							// check if sheep_rx_bytes has been fully read; if so, skip
							if i >= len(sheep_rx_bytes) { 
								done_processing ++
								continue 
							}
							total += int(sheep_rx_bytes[i])
						}
						if done_processing == len(l_m.pasture) - 1 { break }
						rx_bytes_total = append(rx_bytes_total, total)
						i ++
					}
					sort.Ints(rx_bytes_total)
					l_m.rx_bytes_median = rx_bytes_total[len(rx_bytes_total)/2]

					var qps_guess int
					for j := 0; j < len(medians) ; j ++ {
						if l_m.rx_bytes_median >= medians[j] - 3000 {
							if l_m.rx_bytes_median <= medians[j] + 3000 {
								qps_guess = qpses[j]
							}
						}
					}

					fmt.Println("median: ", l_m.rx_bytes_median, "qps guess:", qps_guess)
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







