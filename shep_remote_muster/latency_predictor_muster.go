package main

import (
	"strconv"
)

/**************************************/
func (lat_pred_m *latency_predictor_muster) init() {
	for sheep_id, sheep := range(lat_pred_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		switch {
		case label == "core":
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + lat_pred_m.ip, n_ip: lat_pred_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
			lat_pred_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + lat_pred_m.ip, n_ip: lat_pred_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			lat_pred_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
		default:
		}
	}
}


