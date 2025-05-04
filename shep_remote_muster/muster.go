package main

import (
	"strconv"
	"time"
	"os"
)

//import "fmt"

/*********************************************/
func ctrl_get_remote(core uint8, extra_args ...string) uint64 {
	return 0
}

func ctrl_set_remote(core uint8, val uint64, extra_args ...string) error {
	return nil
}

func (c *control) init(knob string, getter func(uint8, ...string)uint64, setter func(uint8, uint64, ...string)error) {
	c.knob = knob
	c.dirty = false
	c.getter = getter
	c.setter = setter

	c.ready_request_chan = make(chan bool, 1)
	c.ready_ctrl_chan = make(chan bool, 1) // TODO fix; temp

	c.ready_request_chan <- true
}

func (l *log) init_specs(buff_max_size uint64, metrics []string, log_wait_factor time.Duration) {
	mem_buff := make([][]uint64, 0)
	l.metrics = metrics
	l.max_size = buff_max_size
	l.mem_buff = &mem_buff
	l.log_wait_factor = log_wait_factor
}

func (l *log) init() {
	l.kill_log_chan = make(chan bool, 1)
	l.ready_process_chan = make(chan bool, 1)
	l.ready_request_chan = make(chan bool, 1)
	l.ready_buff_chan = make(chan bool, 1)
	l.ready_file_chan = make(chan bool, 1)

	l.ready_process_chan <- true
	l.ready_request_chan <- true
	l.ready_file_chan <- true
}

func (sh *sheep) init() {
	sh.logs = make(map[string]*log)
	sh.controls = make(map[string]*control)
	sh.perf_data = make(map[string][]float32)

	sh.done_ctrl_chan = make(chan control_reply, 1)
	sh.ready_ctrl_chan = make(chan bool, 1)
	sh.new_ctrl_chan = make(chan map[string]uint64)
	sh.request_log_chan = make(chan []string)
	sh.request_ctrl_chan = make(chan string)
	sh.detach_native_logger = make(chan bool, 1)

	// default: minimum of 1 log per-sheep
	log_id := "log-" + sh.label + "-" + strconv.Itoa(int(sh.index)) 
	sh.logs[log_id] = &log{id: log_id}
	sh.logs[log_id].init()
}

func (sheep_c *sheep) write_log_file(log_id string) {
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
}

func (m *muster) init() {
	m.id = "muster-" + m.ip
	if m.node.ip_idx >= 0 {
		ip_idx := strconv.Itoa(m.node.ip_idx)
		m.id = m.id + "-" + ip_idx 
	}

	m.pasture = make(map[string]*sheep)
	m.native_loggers = make(map[string]func(*sheep, *log, string))

	/* log and control synchronization channels */
	m.full_buff_chan = make(chan []string)
	m.process_buff_chan = make(chan []string)
	m.new_ctrl_chan = make(chan control_request)
	/* coordination channels */
	m.request_log_chan = make(chan []string)
	m.request_ctrl_chan = make(chan map[string]map[string]uint64)

	for _, resrc := range(m.resources) {
		sheep_id := "sheep-" + resrc.label + "-" + strconv.Itoa(int(resrc.index)) + "-" + m.ip
		sheep_c := sheep{resource: resrc,
				 id: sheep_id}
		sheep_c.init()
		m.pasture[sheep_id] = &sheep_c
	}

	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	m.logs_dir = home_dir + "/" + m.id + ".logs/" 
}


