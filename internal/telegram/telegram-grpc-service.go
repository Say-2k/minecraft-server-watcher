package telegram

import (
	"log"
	"net"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"

	"google.golang.org/grpc"
)

type BotServer struct {
	agentpb.UnimplementedAgentServiceServer
	botStream        agentpb.AgentService_ConnectServer
	lastStatus       agentpb.ServerStatus
	conf             *config.BotConfig
	telegramNotifier *TelegramNotifier
}

type Status struct {
	grpcStatus agentpb.ServerStatus
	message    string
}

func NewBotServer(conf *config.BotConfig, tn *TelegramNotifier) *BotServer {
	return &BotServer{lastStatus: agentpb.ServerStatus_START, conf: conf, telegramNotifier: tn}
}

func (s *BotServer) Connect(stream agentpb.AgentService_ConnectServer) error {
	log.Println("New gRPC connection established")

	s.botStream = stream
	go s.reciveMessage()

	select {}
}

func (s *BotServer) reciveMessage() error {
	for {
		message, err := s.botStream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return err
		}

		switch message.Status {
		case agentpb.ServerStatus_START:
			log.Println("Received START status")
			s.telegramNotifier.OnStart()
			s.lastStatus = agentpb.ServerStatus_START

		case agentpb.ServerStatus_STOPPING:
			log.Println("Received STOPPING status")
			s.telegramNotifier.OnStop()
			s.lastStatus = agentpb.ServerStatus_STOPPING

		case agentpb.ServerStatus_RUNNING:
			log.Println("Received RUNNING status")
			s.telegramNotifier.OnRunning()
			s.lastStatus = agentpb.ServerStatus_RUNNING
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

	listner, err := net.Listen("tcp", ":"+s.conf.BotPort)
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
