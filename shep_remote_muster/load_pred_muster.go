package main

import (
	"strconv"
)

/*********************************************/

/* a load prediction muster manages one Memcached server
   it holds data to maintain statistics about the load imposed on the server
   it issues control state change requests when that load changes
   the control state managed in this case is the interrupt-delay and dvfs
   setting of the full node 
*/
func (load_pred_m *load_pred_muster) init() {
	load_pred_m.processing_lock = make(chan bool, 1)
	load_pred_m.processing_lock <- true
	
	load_pred_m.rx_bytes_all = make(map[string][]uint64)
	load_pred_m.timestamps_all = make(map[string][]uint64)
	load_pred_m.rx_bytes_concat = make([]uint64, 0)
//	stats_m.rx_bytes_medians = make([]int, 0)

	load_pred_m.ctrl_break = 1
	load_pred_m.cur_load_pred = 0
	load_pred_m.cur_load_guess = 0

	for sheep_id, sheep := range(load_pred_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		switch {
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + load_pred_m.ip, n_ip: load_pred_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + load_pred_m.ip, n_ip: load_pred_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs_all, write_dvfs_all)
			load_pred_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
			load_pred_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		default:
		}
	}
}


