package main

import (
//	"fmt"
	"strconv"
)

/*********************************************/

func (m *muster) init(n_ip string, n_cores int) {
	n := node{ip:n_ip, ncores: uint8(n_cores)}
	m_id := "muster-" + n.ip
	m.id = m_id
	m.node = n 
	m.pasture = make(map[string]*sheep)

	m.hb_chan = make(chan bool)

	m.exit_chan = make(chan bool, 1)
	m.full_buff_chan = make(chan []string)
	m.new_ctrl_chan = make(chan ctrl_req)
	m.done_chan = make(chan []string)

	var core uint8
	for core = 0; core < n.ncores; core ++ {
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + m.ip
		sheep_c := sheep{id: sheep_id, core: core,
				 logs: make(map[string]*log),
				 controls: make(map[string]*control),
				 ready_ctrl_chan: make(chan bool, 1),
				 done_ctrl_chan: make(chan bool, 1),
				 detach_native_logger: make(chan bool, 1),
				 done_kill_chan: make(chan bool, 1)}
		m.pasture[sheep_id] = &sheep_c
	}
}








