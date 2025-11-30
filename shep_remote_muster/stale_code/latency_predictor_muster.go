package main

import (
	"strconv"
	"fmt"
	"os"
	"encoding/csv"
)

/**************************************/
type latency_predictor_muster struct {
	intlog_muster
}

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


func (lat_pred_m *latency_predictor_muster) parse_optimization(opt_req optimize_request) map[string]map[string]uint64 {
	target_ctrls := make(map[string]map[string]uint64)
	for sheep_id, _ := range(lat_pred_m.pasture) {
		target_ctrls[sheep_id] = make(map[string]uint64)
	}

	// as-per specialization: this shepherd expects 1 dvfs setting for all cores 
	// and 1 itr-delay setting for the full node/NIC
	for _, opt_setting := range(opt_req.settings) {
		switch {
		case opt_setting.knob == "itr-delay":
			for _, sheep := range(lat_pred_m.pasture) {
				if sheep.label == "node" {
					target_ctrls[sheep.id]["itr-ctrl-" + sheep.label + "-" + strconv.Itoa(int(sheep.index)) + "-" + lat_pred_m.ip] = opt_setting.val 
					break
				}
			}
		case opt_setting.knob == "dvfs":
			for _, sheep := range(lat_pred_m.pasture) {
				if sheep.label == "core" {
					target_ctrls[sheep.id]["dvfs-ctrl-" + sheep.label + "-" + strconv.Itoa(int(sheep.index)) + "-" + lat_pred_m.ip] = opt_setting.val
				}
			}
		default:
			fmt.Println("****** Unimplemented optimization setting: ", opt_setting.knob)
		}
	}
	return target_ctrls
}


func (lat_pred_m *latency_predictor_muster) init_log_files(opt_req optimize_request) {
	logs_dir := lat_pred_m.logs_dir
	ctrls := opt_req.settings
	for _, ctrl := range(ctrls) {
		logs_dir = logs_dir + ctrl.knob + strconv.Itoa(int(ctrl.val))
	}
	logs_dir = logs_dir + "/"
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }

	for _, sheep := range(lat_pred_m.pasture) {
		sheep.log_f_map = make(map[string]*os.File)
		sheep.log_writer_map = make(map[string]*csv.Writer)
		for log_id, _ := range(sheep.logs) {
			out_fname := logs_dir + log_id
			f, err := os.Create(out_fname)
			if err != nil { panic(err) }
			writer := csv.NewWriter(f)
			writer.Comma = ' '
			sheep.log_f_map[log_id] = f
			sheep.log_writer_map[log_id] = writer
		}
	}
}



