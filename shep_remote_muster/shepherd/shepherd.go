package main

import (
	"fmt"
	"os"
	"strconv"
	"encoding/csv"
	"golang.org/x/exp/maps"
	"slices"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

func (m *muster) init_local(logs_dir string) {
	m.logs_dir = logs_dir
}


/* 
   This function initializes 1) a generic shepherd with a muster representation
   for each node under supervision and 2) a general log and control 
   representation for each core under a muster's supervision. 
*/
func (s *shepherd) init(nodes []node) {
	s.hb_chan = make(chan *pb.HeartbeatReply)
//	s.process_buff_chan = make(chan []string)

	/* init 1 muster for each node and 1 sheep for each core */
	s.musters = make(map[string]*muster)
	s.local_musters = make(map[string]*local_muster)
	for n := 0; n < len(nodes); n++ {
		m_n := muster{node: nodes[n]}
		m_n.init()
		l_m := local_muster{muster: m_n}
		l_m.init()
		s.musters[m_n.id] = &m_n
		s.local_musters[l_m.id] = &l_m
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
		for _, sheep := range(l_m.pasture) {
			sheep.log_f_map = make(map[string]*os.File)
			sheep.log_writer_map = make(map[string]*csv.Writer)
			for log_id, _ := range(sheep.logs) {
				out_fname := logs_dir + log_id
				ctrl_ids := maps.Keys(sheep.controls)
				slices.Sort(ctrl_ids)
				for i := 0; i < len(ctrl_ids); i ++ {
					id := ctrl_ids[i]
					ctrl := sheep.controls[id]
					ctrl_val := strconv.Itoa(int(ctrl.value))
					out_fname += "_" + ctrl_val
				}
				f, err := os.Create(out_fname)
				if err != nil { panic(err) }
				writer := csv.NewWriter(f)
				writer.Comma = ' '
				sheep.log_f_map[log_id] = f
				sheep.log_writer_map[log_id] = writer
			}
		}
	}
}

/* This function starts all muster threads required by a shepherd. */
func (s *shepherd) deploy_musters() {
	for _, l_m := range(s.local_musters) {
		fmt.Printf("\033[97;1m**** DEPLOYING MUSTER %v ****\n\033[0m", l_m.id)
		l_m.start_pulser()		// per-muster pulse client	
		l_m.start_controller()		// per-muster ctrl client
		l_m.start_coordinator()		// per-muster coordinate client
		l_m.start_logger()		// per-muster log server
		go s.log(l_m.id)
		if optimize_on { l_m.start_optimizer() }
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
//			s.process_buff_chan <- []string{m.id, sheep_id, log_id}
			m.process_buff_chan <- []string{sheep_id, log_id}
			go func() {
				m := m
				sheep_id := sheep_id
				log_id := log_id
				<- m.pasture[sheep_id].logs[log_id].ready_process_chan
				select {
				case m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true:
				default:
				}
			} ()
		}
	}
}

func (s *shepherd) control(m_id string, sheep_id string, ctrls map[string]uint64) {
	l_m := s.local_musters[m_id]
	sheep := l_m.pasture[sheep_id]
	l_m.new_ctrl_chan <- control_request{sheep_id: sheep_id, ctrls: ctrls}
	ctrl_reply := <- sheep.done_ctrl_chan
	new_ctrls := ctrl_reply.ctrls
	done_ctrl := ctrl_reply.done
	if done_ctrl {
		for ctrl_id, ctrl_val := range(new_ctrls) {
		        sheep.controls[ctrl_id].value = ctrl_val
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







