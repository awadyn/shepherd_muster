package main

import (
	"strconv"
	"fmt"
//	"os"
//	"encoding/csv"
)

/**************************************/
type bayopt_intlog_muster struct {
	intlog_muster
}

func (bayopt_m *bayopt_intlog_muster) init() {
	for sheep_id, sheep := range(bayopt_m.pasture) {
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		switch {
		case label == "core":
			ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + bayopt_m.ip, n_ip: bayopt_m.ip}
			ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
			bayopt_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
			fmt.Println("SETTER:", bayopt_m.pasture[sheep_id].controls[ctrl_dvfs.id].setter)
		case label == "node":
			ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + bayopt_m.ip, n_ip: bayopt_m.ip}
			ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
			bayopt_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
			fmt.Println("SETTER:", bayopt_m.pasture[sheep_id].controls[ctrl_itr.id].setter)
		default:
		}
	}
}




