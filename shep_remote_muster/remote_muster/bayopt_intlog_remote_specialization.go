package main

import (
	"fmt"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

func (bayopt_m *bayopt_intlog_muster) init_remote() {
	iface := get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: %v cannot get internal network interface name.. aborting\n", bayopt_m.id)
		return
	}
	for _, sheep := range(bayopt_m.pasture) {
		for _, ctrl := range(sheep.controls) {
			switch {
			case ctrl.knob == "itr-delay":
				rx_usecs_reading := ctrl.getter(0, iface)	
				if rx_usecs_reading == 0 {
					fmt.Printf("**** PROBLEM: %v cannot read ethtool rx_usecs value.. assuming default value 1\n", bayopt_m.id)
					ctrl.value = 1
				} else {
					ctrl.value = rx_usecs_reading
				}
			case ctrl.knob == "dvfs":
				dvfs_reading := ctrl.getter(sheep.core)
				if dvfs_reading == 0 {
					fmt.Printf("**** PROBLEM: %v cannot read dvfs value.. assuming default value 0xffff\n", bayopt_m.id)
					ctrl.value = 0xffff
				} else {
					ctrl.value = dvfs_reading
				}
			default:
				fmt.Printf("**** PROBLEM: UNKOWN CTRL KNOB -- %v - %v - %v\n", bayopt_m.id, sheep.id, ctrl.id)
			}
		}
	}
}


//func (intlog_m *intlog_muster) ctrl_manage(sheep_id string) {
//	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", intlog_m.id, sheep_id)
//	sheep := intlog_m.pasture[sheep_id]
//	var err error
//	for {
//		select {
//		case new_ctrls := <- sheep.new_ctrl_chan:
//			for ctrl_id, ctrl_val := range(new_ctrls) {
//				switch {
//				case sheep.controls[ctrl_id].knob == "dvfs":
//					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
//				case sheep.controls[ctrl_id].knob == "itr-delay":
//					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
//				default:
//				}
//				if err != nil { panic(err) }
//				sheep.controls[ctrl_id].value = ctrl_val
//			}
//			intlog_m.pasture[sheep.id].done_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
//		}
//	}
//}



