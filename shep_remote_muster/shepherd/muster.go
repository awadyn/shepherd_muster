package main

import (
	"strconv"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/*********************************************/

func (m *muster) init() {
	m.hb_chan = make(chan *pb.HeartbeatReply)
	m.full_buff_chan = make(chan []string)
	m.new_ctrl_chan = make(chan control_request)

	m.pasture = make(map[string]*sheep)
	var c uint8
	for c = 0; c < m.ncores; c++ {
		sheep_id := strconv.Itoa(int(c)) + "-" + m.ip
		sheep_c := sheep{id: sheep_id, core: c,
				 logs: make(map[string]*log), 
				 controls: make(map[string]*control),
				 done_ctrl_chan: make(chan control_reply, 1),
				 finish_run_chan: make(chan bool, 1)}
		m.pasture[sheep_id] = &sheep_c
	}
}

