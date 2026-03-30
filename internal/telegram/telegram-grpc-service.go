package telegram

import (
	"log"

	agentpb "github.com/Say-2k/minecraft-server-watcher/v2/api/agent/v1/agentpb"
)

type BotServer struct {
	agentpb.UnimplementedAgentServiceServer
}

func NewBotServer() *BotServer {
	return &BotServer{}
}

func (s *BotServer) Connect(stream agentpb.AgentService_ConnectServer) error {
	log.Println("New gRPC connection established")

}
