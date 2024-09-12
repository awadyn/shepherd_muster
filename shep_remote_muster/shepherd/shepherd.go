package main

import (
	"fmt"
	"os"
	"strconv"
	"encoding/csv"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

/* 
   This function initializes 1) a generic shepherd with a muster representation
   for each node under supervision and 2) a general log and control 
   representation for each core under a muster's supervision. 
*/
func (s *shepherd) init(nodes []node) {
	s.hb_chan = make(chan *pb.HeartbeatReply)
	s.process_buff_chan = make(chan []string)
	s.compute_ctrl_chan = make(chan []string)

	/* init 1 muster for each node and 1 sheep for each core */
	s.musters = make(map[string]*muster)
	s.local_musters = make(map[string]*local_muster)
	for n := 0; n < len(nodes); n++ {
		m_n := &muster{node: nodes[n]}
		m_n.init()
		l_m := &local_muster{muster: *m_n}
		l_m.init()
		s.musters[m_n.id] = m_n
		s.local_musters[l_m.id] = l_m
	}
}

/* This function assigns a map of log files to each sheep/core
   such that there can then be a separate per-sheep log for different 
   control settings.
*/
func (s *shepherd) init_log_files(logs_dir string) {
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	for _, l_m := range(s.local_musters) {

		l_m.out_f_map = make(map[string](map[string]*os.File))
		l_m.out_writer_map = make(map[string](map[string]*csv.Writer))
		l_m.out_f = make(map[string]*os.File)
		l_m.out_writer = make(map[string]*csv.Writer)
		for _, sheep := range(l_m.pasture) {
			l_m.out_f_map[sheep.id] = make(map[string]*os.File)
			l_m.out_writer_map[sheep.id] = make(map[string]*csv.Writer)
			c_str := strconv.Itoa(int(sheep.core))
			out_fname := logs_dir + l_m.id + "_" + c_str 
			for _, ctrl := range(sheep.controls) {
				ctrl_val := strconv.Itoa(int(ctrl.value))
				out_fname += "_" + ctrl_val
			}
			f, err := os.Create(out_fname)
			if err != nil { panic(err) }
			writer := csv.NewWriter(f)
			writer.Comma = ' '
			l_m.out_f_map[sheep.id][out_fname] = f
			l_m.out_writer_map[sheep.id][out_fname] = writer
			l_m.out_f[sheep.id] = f
			l_m.out_writer[sheep.id] = writer
		}
	}
}

/* This function starts all muster threads required by a shepherd. */
func (s *shepherd) deploy_musters() {
	for _, l_m := range(s.local_musters) {
		fmt.Printf("\033[97;1m**** DEPLOYING MUSTER %v ****\n\033[0m", l_m.id)
		l_m.start_pulser()		// per-muster pulse client	
//		l_m.start_controller()		// per-muster ctrl client
		l_m.start_coordinator()		// per-muster coordinate client
		l_m.start_logger()		// per-muster log server
		go s.log(l_m.id)
	}
}

/********** PULSING **********/

/* shepherd pulsing: handles pulse signals received
   on shepherd's heartbeat channel. This channel is unbuffered.
   It receives a message each time a muster pulses.
*/
func (s *shepherd) listen_heartbeats() {
	fmt.Printf("\033[39;1m-- STARTING HEARTBEAT LISTENER :  %v\n\033[0m", s.id)
	counter := 0
	for {
		for _, m := range(s.local_musters) {
			select {
			case r := <- s.local_musters[m.id].hb_chan:
				m_id := r.GetMusterReply()
				if counter % 3 == 0 { fmt.Printf("\033[39m-- HB REP %v - %v\n\033[0m", r.GetShepRequest(), m_id) }
			default:
			}
			counter ++
		}
	}
}


/********** LOGGING **********/

/* shepherd logging: handles log management signals received 
   on a muster's full_buff channel. This channel is unbuffered.
   It receives a message each time a sheep requires log management.
*/
func (s *shepherd) log(m_id string) {
	m := s.musters[m_id]
	for {
		select {
		case ids := <- m.full_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			s.process_buff_chan <- []string{m.id, sheep_id, log_id}
			go func() {
				<- m.pasture[sheep_id].logs[log_id].ready_process_chan
				select {
				case m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true:
				default:
				}
			} ()
		}
	}
}



//func (s *shepherd) CompleteRun(ctx context.Context, in *pb.CompleteRunRequest) (*pb.CompleteRunReply, error) {
//	sheep_id := in.GetSheepId()
//	muster_id := in.GetMusterId()
//	s.complete_run_chan <- []string{muster_id, sheep_id}
//	<- s.musters[muster_id].pasture[sheep_id].finish_run_chan
//	return &pb.CompleteRunReply{RunComplete: true}, nil
//}
//
//func (s *shepherd) complete_runs() {
//	flag.Parse()
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *s.coordinate_port))
//	if err != nil {
//		fmt.Printf("** ** ** ERROR: %v failed to listen at %v: %v\n", s.id, *s.coordinate_port, err)
//	}
//	server := grpc.NewServer()
//	pb.RegisterCoordinateServer(server, s)
//	fmt.Printf("-- %v -- Coordination server listening at %v ... ... ...\n", s.id, lis.Addr())
//	if err := server.Serve(lis); err != nil {
//		fmt.Printf("** ** ** ERROR: %v failed to start coordination server: %v\n", s.id, err)
//	}
//}







