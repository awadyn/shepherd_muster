package main

import (
	"fmt"
	"os"
	"encoding/csv"
	"time"
	
	"context"
	"flag"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	pb_opt "github.com/awadyn/shep_remote_muster/shep_log_processor"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
	s.new_ctrl_chan = make(chan map[string]control_request)

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
	for _, l_m := range(s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			sheep.log_f_map = make(map[string]*os.File)
			sheep.log_writer_map = make(map[string]*csv.Writer)
			for log_id, _ := range(sheep.logs) {
				out_fname := logs_dir + log_id
//				ctrl_ids := maps.Keys(sheep.controls)
//				slices.Sort(ctrl_ids)
//				for i := 0; i < len(ctrl_ids); i ++ {
//					id := ctrl_ids[i]
//					ctrl := sheep.controls[id]
//					ctrl_val := strconv.Itoa(int(ctrl.value))
//					out_fname += "_" + ctrl_val
//				}
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

		go s.process_logs(l_m.id)
		go s.process_control(l_m.id)

		if optimize_on { 
			for id, _ := range(s.optimizers) {
				s.start_optimizer_server(id)
				s.start_optimizer_client(id)
			}
		}

		l_m.show()
	}
	go s.listen_heartbeats()
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
		time.Sleep(time.Second)
	}
}


/********** LOGGING **********/

/* shepherd logging: handles log management signals received 
   on a muster's full_buff channel. This channel is unbuffered.
   It receives a message each time a sheep requires log syncing.
*/
func (s shepherd) process_logs(m_id string) {
	l_m := s.local_musters[m_id]
	for {
		select {
		case ids := <- l_m.full_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]

			go func() {
				l_m := l_m
				sheep_id := sheep_id
				log_id := log_id
				sheep := l_m.pasture[sheep_id]
				log := sheep.logs[log_id]

				if debug { fmt.Printf("\033[32m-------- PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id) }
				sheep.write_log_file(log_id)

				if specialize_on {
					go func() {
						l_m := l_m
						log := log
						sheep_id := sheep_id
						log_id := log_id

						l_m.process_buff_chan <- []string{sheep_id, log_id}
						<- log.ready_process_chan
					} ()
				}

				if debug { fmt.Printf("\033[32m-------- COMPLETED PROCESS LOG :  %v - %v\n\033[0m", sheep_id, log_id) }
				log.ready_buff_chan <- true // muster can now overwrite log mem_buff
			} ()
		}
	}
}


func (s shepherd) process_control(m_id string) {
	l_m := s.local_musters[m_id]
	for {
		select {
		case target_ctrls := <- l_m.request_ctrl_chan:
			for sheep_id, ctrls := range(target_ctrls) {
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
			
			select {
			case l_m.done_request_chan <- true:
			default:
			}
		}
	}		
}



// set controls remotely, when done, set locally
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

/*
func (s shepherd) compute_control(m_id string, ctrl_parser func([]optimize_setting)map[string]uint64) {
	l_m := s.local_musters[m_id]
	for {
		select {
		case opt_req := <- l_m.request_optimize_chan:
			fmt.Printf("\033[31m-------- REQUEST OPTIMIZE SIGNAL :  %v - %v\n\033[0m", m_id, opt_req)
			opt_settings := opt_req.settings
			ctrls := ctrl_parser(opt_settings)
			l_m.ready_optimize_chan <- ctrls
		}
	}		
}

func (s *shepherd) CompleteRun(ctx context.Context, in *pb.CompleteRunRequest) (*pb.CompleteRunReply, error) {
	sheep_id := in.GetSheepId()
	muster_id := in.GetMusterId()
	s.complete_run_chan <- []string{muster_id, sheep_id}
	<- s.musters[muster_id].pasture[sheep_id].finish_run_chan
	return &pb.CompleteRunReply{RunComplete: true}, nil
}

func (s *shepherd) complete_runs() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *s.coordinate_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen at %v: %v\n", s.id, *s.coordinate_port, err)
	}
	server := grpc.NewServer()
	pb.RegisterCoordinateServer(server, s)
	fmt.Printf("-- %v -- Coordination server listening at %v ... ... ...\n", s.id, lis.Addr())
	if err := server.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to start coordination server: %v\n", s.id, err)
	}
}
*/


/********** OPTIMIZATION **********/

func (s *shepherd) start_optimizer_server(id string) {
	fmt.Printf("\033[34;1m-- %v: STARTING OPTIMIZER :  %v\n\033[0m", s.id, id)
	port := s.optimizers[id].port
	go s.optimize_server(port)
}

func (s *shepherd) optimize_server(port *int) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to listen at %v: %v\n\033[0m", s.id, *port, err)
	}
	serve := grpc.NewServer()
	pb_opt.RegisterLogStatsMessengerServer(serve, s)
	fmt.Printf("\033[36m---- %v -- Initialized optimization server listening at %v \n\033[0m", s.id, lis.Addr())
	if err := serve.Serve(lis); err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to start optimization server: %v\n\033[0m", s.id, err)
	}
}

func (s *shepherd) EvaluateLogStats(ctx context.Context, in *pb_opt.EvaluateLogStatsRequest) (*pb_opt.EvaluateLogStatsReply, error) {
	if debug { fmt.Printf("\033[0m-----> OPTIMIZE-REQ -- %v - ARGS -- %v\n\033[0m", s.id, in.GetArgs()) }
	for _, arg := range(in.GetArgs()) {
		value := wrapperspb.Int64(0)
		if err := arg.UnmarshalTo(value); err != nil { 
			fmt.Println("ERROR : ", err) 
		}
		fmt.Println(value) 
	}
	return &pb_opt.EvaluateLogStatsReply{Done: true}, nil
}

func (s *shepherd) start_optimizer_client(id string) {
	addr := s.optimizers[id].addr
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("****** ERROR: %v could not create connection to optimizer server:\n****** %v\n", s.id, err)
	}
	c := pb_opt.NewLogProcessorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("\033[36m---- %v -- Initialized optimizer client\n\033[0m", s.id)
	s.optimize_client(conn, c, ctx, cancel)
}

func (s *shepherd) optimize_client(conn *grpc.ClientConn, c pb_opt.LogProcessorClient, ctx context.Context, cancel context.CancelFunc) {
	for id, optimizer := range(s.optimizers) {
		go func() {
			conn := conn
			c := c
			ctx := ctx
			cancel := cancel
			id := id
			optimizer := optimizer
			defer conn.Close()
			defer cancel()
			for {
				select {
				case start_req := <- optimizer.start_optimize_chan:
					args := start_req.args
					//fmt.Println("***************** start_req: ", start_req, " -- args: ", args)
					r, err := c.StartOptimizer(ctx, &pb_opt.StartOptimizerRequest{Args: args})  
					if err != nil {
						fmt.Printf("\033[31;1m***** COULD NOT START OPTIMIZER %v:  %v\n\033[0m", id, err)
						optimizer.ready_optimize_chan <- r.GetDone()
						return
					} else {
						fmt.Printf("\033[34;1m***** STARTED OPTIMIZER:  %v - %v\n\033[0m", id, r.GetDone())
						optimizer.ready_optimize_chan <- r.GetDone()
					}
				case opt_req := <- optimizer.request_optimize_chan:
					args := opt_req.args
					//fmt.Println("***************** opt_req: ", opt_req, " -- args: ", args)
					r, err := c.ProcessLog(ctx, &pb_opt.ProcessLogRequest{Args: args})  
					if err != nil {
						fmt.Printf("\033[31;1m***** COULD NOT PROCESS LOG %v:  %v\n\033[0m", id, err)
					} else {
						if debug { fmt.Printf("\033[34;1m***** DONE PROCESS LOG:  %v - %v\n\033[0m", id, r.GetDone()) }
					}
					optimizer.done_optimize_chan <- r.GetDone()
				}
			}
		}()
	}
}


/********** CLEANUP **********/

func (s *shepherd) cleanup() {
	for _, l_m := range(s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}


