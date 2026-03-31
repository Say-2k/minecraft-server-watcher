package telegram

import (
	"log"
	"net"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"

	"google.golang.org/grpc"
)

type BotServer struct {
	agentpb.UnimplementedAgentServiceServer
	botStream  agentpb.AgentService_ConnectServer
	lastStatus agentpb.ServerStatus
}

type Status struct {
	grpcStatus agentpb.ServerStatus
	message    string
}

func NewBotServer() *BotServer {
	return &BotServer{}
}

func (s *BotServer) Connect(stream agentpb.AgentService_ConnectServer) error {
	log.Println("New gRPC connection established")

	go reciveMessage(stream)

	select {}
}

func reciveMessage(stream agentpb.AgentService_ConnectServer) error {
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return err
		}

		switch message.Status {
		case agentpb.ServerStatus_START:
			log.Println("Received START status")
		case agentpb.ServerStatus_STOPPING:
			log.Println("Received STOPPING status")
		case agentpb.ServerStatus_RUNNING:
			log.Println("Received RUNNING status")
		}
	}
}

func (s *BotServer) sendMessage(msg agentpb.Command) error {
	if err := s.botStream.Send(&agentpb.BotMessage{Command: msg}); err != nil {
		log.Printf("Error sending message: %v", err)
		return err
	}

	return nil
}

func (s *BotServer) Start() error {
	log.Println("Start telegram bot server")

	listner, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	agentpb.RegisterAgentServiceServer(grpcServer, s)
	if err := grpcServer.Serve(listner); err != nil {
		log.Printf("Failed to serve: %v", err)
		return err
	}

	return nil
}
