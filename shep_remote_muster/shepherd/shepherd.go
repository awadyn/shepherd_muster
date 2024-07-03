package main

import (
	"fmt"

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
	//s.compute_ctrl_chan = make(chan []string)
	//s.complete_run_chan = make(chan []string)
	//s.coordinate_port = flag.Int("coordinate_port", 50081,
	//			     "shepherd coordination server port")

	/* init 1 muster for each node and 1 sheep for each core */
	s.musters = make(map[string]*muster)
	s.local_musters = make(map[string]*local_muster)
	for n := 0; n < len(nodes); n++ {
		m_id := "muster-" + nodes[n].ip
		m_n := &muster{id: m_id, node: nodes[n]}
		m_n.init()
		l_m := &local_muster{muster: *m_n}
		l_m.init()
		s.musters[m_id] = m_n
		s.local_musters[m_id] = l_m
	}
}


/* This function starts all threads required by a shepherd 
   to manage all of its musters and their sheep.
*/
func (s *shepherd) deploy_musters() {
	for _, l_m := range(s.local_musters) {	
		l_m.start_pulser()		// general functionality: start pulse management thread per-muster
		l_m.start_logger()		// * specialization: start log server
		go s.log(l_m.id)		// general functionality: start log management thread per-muster
//		go l_m.start_controller()
	}
}

/********** PULSING **********/

/* shepherd pulsing: handles pulse signals received
   on shepherd's heartbeat channel. This channel is unbuffered.
   It receives a message each time a muster pulses.
*/
func (s *shepherd) listen_heartbeats() {
	fmt.Printf("-- STARTING HEARTBEAT LISTENER :  %v ... ... ...\n", s.id)
	for {
		for _, m := range(s.musters) {
			select {
			case r := <- s.musters[m.id].hb_chan:
				m_id := r.GetMusterReply()
				fmt.Println("------HB-REP --", m_id, r.GetShepRequest())
//				s.musters[m_id].pulsing = true
			default:
			}
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
				<- m.pasture[sheep_id].logs[log_id].done_process_chan
				select {
				case m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true:
				default:
				}
			} ()
		}
	}
}


//func (s *shepherd) control(m_id string) {
//	m := s.musters[m_id]
//	for {
//		select {
//		case new_ctrl_req := <- m.new_ctrl_chan:
//			sheep_id := new_ctrl_req.sheep_id
//			new_ctrls := new_ctrl_req.ctrls
//			
//			
//		}
//	}
//}

/**/

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







