package main

import (
	"strconv"
//	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/*********************************************/

func (m *muster) init() {
	/* log and control synchronization channels */
	m.full_buff_chan = make(chan []string)
	m.new_ctrl_chan = make(chan control_request)

	/* coordination channels */
	m.request_log_chan = make(chan []string)
	m.request_ctrl_chan = make(chan []string)

	m.pasture = make(map[string]*sheep)
	var c uint8
	for c = 0; c < m.ncores; c++ {
		sheep_id := "sheep-" + strconv.Itoa(int(c)) + "-" + m.ip
		sheep_c := sheep{id: sheep_id, core: c,
				 logs: make(map[string]*log), 
				 controls: make(map[string]*control),
//				 done_request_chan: make(chan bool, 1),
				 done_ctrl_chan: make(chan control_reply, 1)}
		m.pasture[sheep_id] = &sheep_c
	}
}

