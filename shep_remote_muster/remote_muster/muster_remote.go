package main

import (
//	"fmt"
//	"time"
	"io"
	"strconv"
	"encoding/csv"
	"os"
)

/*********************************************/

func (m *muster) init_remote() {
	for _, sheep := range(m.pasture) {
		sheep.new_ctrl_chan = make(chan map[string]uint64)
		sheep.request_log_chan = make(chan []string)
		sheep.request_ctrl_chan = make(chan string)
		sheep.detach_native_logger = make(chan bool, 1)
		//sheep.done_kill_chan = make(chan bool, 1)
	}
}

func (m *muster) init_log_files(logs_dir string) {
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
//	m.log_f_map = make(map[string]*os.File)
//	m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, sheep := range(m.pasture) {
		sheep.log_f_map = make(map[string]*os.File)
		sheep.log_reader_map = make(map[string]*csv.Reader)
		
		core := m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := logs_dir + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }

		reader := csv.NewReader(f)
		reader.Comma = ' '
		for _, log := range(sheep.logs) {
			sheep.log_f_map[log.id] = f
			sheep.log_reader_map[log.id] = reader
		}
//		m.log_f_map[sheep_id] = f
	}
}

func (m *muster) cleanup() {
	for _, sheep := range(m.pasture) {
//		m.log_f_map[sheep_id].Close()
		for _, f := range(sheep.log_f_map) {
			f.Close()
		}
	}
}

//func (ctrl *control) getter(core uint8, get_func func(uint8, ...string)uint64, args ...string) uint64 {
//	var ctrl_val uint64
//	if len(args) > 0 {
//		ctrl_val = get_func(core, args[0])
//	} else {
//		ctrl_val = get_func(core)
//	}
//	return ctrl_val
//}
//func (ctrl *control) setter(core uint8, value uint64, set_func func(uint8, uint64)error) error {
//	err := set_func(core, value)
//	return err
//}

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
		err = logger_func(shared_log, reader)
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
			sheep := m.pasture[new_ctrl_req.sheep_id]
			new_ctrls := new_ctrl_req.ctrls
			sheep.new_ctrl_chan <- new_ctrls
		}
	}
}








