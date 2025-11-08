package main

import (
	"strconv"
)

/*********************************************/
func setup_target_resources_c8220(ncores uint8) []resource {
	target_resources := make([]resource, 0)
	var i uint8
	for i = 0; i < ncores; i++ {
		if i < 9 { target_resources = append(target_resources, resource{label: "core", index: i}) }
		if i > 9 { target_resources = append(target_resources, resource{label: "core", index: i}) }
	}
	target_resources = append(target_resources, resource{label: "node", index: 0})
	return target_resources
}

func (stats_m *stats_muster) init() {
	stats_m.rx_bytes_all = make(map[string][]uint64)
	stats_m.timestamps_all = make(map[string][]uint64)
	stats_m.rx_bytes_concat = make([]int, 0)
	stats_m.rx_bytes_medians = make([]int, 0)
	stats_m.processing_lock = make(chan bool, 1)
	stats_m.processing_lock <- true

	for sheep_id, sheep := range(stats_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		switch {
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + stats_m.ip, n_ip: stats_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + stats_m.ip, n_ip: stats_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs_all, write_dvfs_all)
			stats_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
			stats_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		default:
		}
	}
}

