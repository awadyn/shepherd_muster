package main

import (
	"fmt"
	"context"
	"flag"
	"net"
	"io"
	"time"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

func (l_m *local_muster) start_pulser(conn *grpc.ClientConn, c pb.PulseClient, ctx context.Context, cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()

	var counter uint32 = 0
	for {
		counter += 1
		r, err := c.HeartBeat(ctx, &pb.HeartbeatRequest{ShepRequest: counter})  
		if err != nil {
//			fmt.Printf("** ** ** ERROR: %v could not send heartbeat request to remote:\n** ** ** %v\n", l_m.id, err)
			time.Sleep(time.Second/5)
			continue
		}
		select {
		case l_m.hb_chan <- r:
//		default:
		}
		time.Sleep(time.Second/5)
	}
	
}

func (l_m *local_muster) SyncLogBuffers(stream pb.Log_SyncLogBuffersServer) error {
	<- l_m.hb_chan

	l_buff_ctr := 0
	var log_id string
	var log_ptr *log
	for {
		log_sync_req, err := stream.Recv()
		switch {
		case err == io.EOF:
			fmt.Printf("------------COMPLETED-SYNC-REQ-- %v -- %v\n", l_m.id, log_id)
			// signal shepherd to start processing synced log
			l_m.process_chan <- log_id
			// signal remote muster OK to flush remote buff
			return stream.SendAndClose(&pb.SyncLogReply{SyncComplete:true})
		case err != nil:
			fmt.Printf("** ** ** ERROR: could not receive sync log request from stream\n")
			// signal remote muster that sync request has failed
			return stream.SendAndClose(&pb.SyncLogReply{SyncComplete:false})
//			panic(err)
//			return err
		default:
			log_id = log_sync_req.GetLogId()
			log_ptr = l_m.logs[log_id]
			buff_ptr := log_ptr.l_buff
			if l_buff_ctr == 0 { 
				fmt.Printf("------------SYNC-REQ-- %v\n", log_id) 
				// confirm that requested sync local log buffer is not in use by shepherd's log processing thread
				<- log_ptr.ready_chan
			}	
			// copy sync log data into local log memory buffer 
			(*buff_ptr)[l_buff_ctr] = log_sync_req.GetLogEntry().GetVals()
			l_buff_ctr++
		}
	}
}

/* This function starts a local muster thread that serves
   log sync requests from its mirror remote muster. A remote
   muster sends a log sync request whenever one of its log
   memory buffers are full. A remote muster can only continue
   logging into its memory buffers after log syncing with 
   the local muster is complete.
*/
func (l_m *local_muster) log() {
	// begin logging protocol once heartbeats are established
	<- l_m.hb_chan

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *l_m.log_sync_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen at %v: %v\n", l_m.id, *l_m.log_sync_port, err)
	}
	fmt.Printf("-- %v -- listening at %v\n", l_m.id, lis.Addr())
	s := grpc.NewServer()
	pb.RegisterLogServer(s, l_m)
	fmt.Printf("-- %v -- starting log sync server\n", l_m.id)
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to start log sync server: %v\n", l_m.id, err)
	}
}

func (l_m *local_muster) control() {
}


