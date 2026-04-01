package telegram

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"

	"google.golang.org/grpc"
)

type BotServer struct {
	agentpb.UnimplementedAgentServiceServer
	botStream        agentpb.AgentService_ConnectServer
	lastStatus       agentpb.ServerStatus
	conf             *config.BotConfig
	telegramNotifier ITelegramNotifier
	mutex            *sync.Mutex
	ctxCancel        context.CancelFunc
}

type IBotServer interface {
	Connect(agentpb.AgentService_ConnectServer) error
	Start() error
	Disconnect()
	SendMessage(agentpb.Command) error
	GetLastStatus() agentpb.ServerStatus
}

type Status struct {
	grpcStatus agentpb.ServerStatus
	message    string
}

func NewBotServer(conf *config.BotConfig, tn ITelegramNotifier) (IBotServer, error) {
	if conf == nil {
		log.Println("BotConfig is nil")
		return nil, fmt.Errorf("BotConfig is nil")
	}
	if tn == nil {
		log.Println("TelegramNotifier is nil")
		return nil, fmt.Errorf("TelegramNotifier is nil")
	}
	return &BotServer{lastStatus: agentpb.ServerStatus_STOPPING, conf: conf, telegramNotifier: tn, mutex: &sync.Mutex{}}, nil
}

func (t *BotServer) GetLastStatus() agentpb.ServerStatus {
	if t == nil || t.mutex == nil {
		return agentpb.ServerStatus_STOPPING
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.lastStatus
}

func (s *BotServer) Start() error {
	log.Println("Start telegram bot server")

	listener, err := net.Listen("tcp", ":"+s.conf.BotPort)
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	agentpb.RegisterAgentServiceServer(grpcServer, s)
	if err := grpcServer.Serve(listener); err != nil {
		log.Printf("Failed to serve: %v", err)
		return err
	}

	return nil
}

func (s *BotServer) Connect(stream agentpb.AgentService_ConnectServer) error {
	log.Println("New gRPC connection established")

	ctx, cancel := context.WithCancel(context.Background())
	s.ctxCancel = cancel
	s.botStream = stream
	go s.reciveMessage()

	select {
	case <-stream.Context().Done():
		log.Println("gRPC connection closed")
		return nil
	case <-ctx.Done():
		log.Println("Context cancelled, closing gRPC connection")
		return nil
	}
}

func (s *BotServer) Disconnect() {
	if s.ctxCancel != nil {
		s.ctxCancel()
	}
}

func (s *BotServer) SendMessage(msg agentpb.Command) error {
	if s.botStream == nil {
		log.Println("Bot stream is not initialized")
		return fmt.Errorf("Bot stream is not initialized")
	}

	if err := s.botStream.Send(&agentpb.BotMessage{Command: msg}); err != nil {
		log.Printf("Error sending message: %v", err)
		return err
	}

	return nil
}

func (s *BotServer) reciveMessage() error {
	for {
		message, err := s.botStream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return err
		}

		s.mutex.Lock()
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
		s.mutex.Unlock()
	}
}
