package main

import (
	"fmt"
//	"strconv"
//	"encoding/csv"
//	"time"
//	"io"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

type nop_muster struct {
	remote_muster
	logs_dir string
}

func (nop_m *nop_muster) init_remote() {
	nop_m.init_log_files(nop_m.logs_dir)
}


func (nop_m *nop_muster) assign_log_files(sheep_id string) {
}


func nop_native_log(sheep *sheep, log *log, logs_dir string) {
}

func (nop_m *nop_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", nop_m.id, sheep_id)
	sheep := nop_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.new_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			nop_m.pasture[sheep.id].done_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
		}
	}
}







