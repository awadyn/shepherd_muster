package main

import (
	"strconv"
	"time"
)

/*********************************************/
func (c *control) init(knob string, getter func(uint8, ...string)uint64, setter func(uint8, uint64)error) {
	c.knob = knob
	c.dirty = false
	c.ready_request_chan = make(chan bool, 1)
	c.getter = getter
	c.setter = setter
}

func (l *log) init(buff_max_size uint64, metrics []string, log_wait_factor time.Duration) {
	mem_buff := make([][]uint64, buff_max_size)
	l.metrics = metrics
	l.max_size = buff_max_size
	l.mem_buff = &mem_buff
	l.log_wait_factor = log_wait_factor
	l.ready_process_chan = make(chan bool, 1)
	l.ready_request_chan = make(chan bool, 1)
	l.ready_buff_chan = make(chan bool, 1)
//	l.kill_log_chan = make(chan bool, 1)
}

func (sheep_c *sheep) update_log_file(log_id string) {
	str_mem_buff := make([][]string,0)
	log := sheep_c.logs[log_id]
	mem_buff := *(log.mem_buff)
	for _, row := range(mem_buff) {
		if len(row) == 0 { break }
		str_row := []string{}
		for i := range(len(log.metrics)) {
			val := strconv.Itoa(int(row[i]))
			str_row = append(str_row, val)
		}
		str_mem_buff = append(str_mem_buff, str_row)
	}
	writer := sheep_c.log_writer_map[log_id]
	writer.WriteAll(str_mem_buff)
	fmt.Println(mem_buff)
	fmt.Println(str_mem_buff)
}

func (m *muster) init() {
	m.id = "muster-" + m.ip
	if m.node.ip_idx >= 0 {
		ip_idx := strconv.Itoa(m.node.ip_idx)
		m.id = m.id + "-" + ip_idx 
	}

	m.pasture = make(map[string]*sheep)

	/* log and control synchronization channels */
	m.full_buff_chan = make(chan []string)
	m.new_ctrl_chan = make(chan control_request)
	/* coordination channels */
	m.request_log_chan = make(chan []string)
	m.request_ctrl_chan = make(chan []string)

	var c uint8
	for c = 0; c < m.ncores; c++ {
		sheep_id := "sheep-" + strconv.Itoa(int(c)) + "-" + m.ip
		sheep_c := sheep{id: sheep_id, core: c,
				 logs: make(map[string]*log), 
				 controls: make(map[string]*control),
				 ready_ctrl_chan: make(chan control_reply, 1)}
		m.pasture[sheep_id] = &sheep_c
	}
}


