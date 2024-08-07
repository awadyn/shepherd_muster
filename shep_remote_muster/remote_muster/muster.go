package main

import (
//	"fmt"
	"io"
	"strconv"
	"encoding/csv"
	"os"
)

/*********************************************/

func (l *log) init() {
	l.kill_log_chan = make(chan bool, 1)
	l.request_log_chan = make(chan string)
	l.done_log_chan = make(chan bool, 1)
	l.ready_buff_chan = make(chan bool, 1)
}

func (m *muster) init(n_ip string, n_cores uint8) {
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
				 request_ctrl_chan: make(chan map[string]uint64),
				 ready_ctrl_chan: make(chan bool, 1),
				 done_ctrl_chan: make(chan bool, 1),
				 detach_native_logger: make(chan bool, 1),
				 done_kill_chan: make(chan bool, 1)}
		m.pasture[sheep_id] = &sheep_c
	}
}


func (m *muster) init_log_files(logs_dir string) {
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	m.log_f_map = make(map[string]*os.File)
	m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, _ := range(m.pasture) {
		core := m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := logs_dir + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		m.log_f_map[sheep_id] = f
	}
}

func (m *muster) cleanup() {
	for sheep_id, _ := range(m.pasture) {
		m.log_f_map[sheep_id].Close()
	}
}

func (ctrl *control) getter(core uint8, get_func func(uint8)uint64) uint64 {
	ctrl_val := get_func(core)
	return ctrl_val
}

func (ctrl *control) setter(core uint8, value uint64, set_func func(uint8, uint64)error) error {
	err := set_func(core, value)
	return err
}

func (m *muster) sync_with_logger(sheep_id string, log_id string, reader *csv.Reader, logger_func func(*log, *csv.Reader)error, n_iter int) error {
	shared_log := m.pasture[sheep_id].logs[log_id] 
	var err error = nil
	iter := 0
	for {
		switch {
		case n_iter > 0:
			if iter == n_iter { return err }
			iter ++
		default:
		}
		err := logger_func(shared_log, reader)
		switch {
		case err == nil:	// => logged one buff
			m.full_buff_chan <- []string{sheep_id, log_id}
			<- shared_log.ready_buff_chan
			*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
		case err == io.EOF:	// => log reader at EOF
			// do nothing if nothing has been logged yet
			if len(*(shared_log.mem_buff)) == 0 { return io.EOF }
			// otherwise sync whatever has been logged with mirror
			m.full_buff_chan <- []string{sheep_id, log_id}
			<- shared_log.ready_buff_chan
			return io.EOF
		}
	}
}

func (m *muster) sync_new_ctrl() {
	for {
		select {
		case new_ctrl_req := <- m.new_ctrl_chan:
			sheep_id := new_ctrl_req.sheep_id
			new_ctrls := new_ctrl_req.ctrls
			m.pasture[sheep_id].request_ctrl_chan <- new_ctrls
//			sheep := m.pasture[sheep_id]
//			for ctrl_id, ctrl_val := range(new_ctrls) {
//				sheep.controls[ctrl_id].value = ctrl_val
//			}
//			m.pasture[sheep.id].ready_ctrl_chan <- true
		}
	}
}








