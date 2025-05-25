package main

import (
	"context"
	agent "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent"
)

// AgentImpl implements the last service interface defined in the IDL.
type AgentImpl struct{}

// Ask implements the AgentImpl interface.
func (s *AgentImpl) Ask(ctx context.Context, askRequest *agent.AskRequest) (resp *agent.AskResponse, err error) {
	// TODO: Your code here...
	return
}

// StartSummary implements the AgentImpl interface.
func (s *AgentImpl) StartSummary(ctx context.Context, summaryRequest *agent.StartSummaryRequest) (resp *agent.StartSummaryResponse, err error) {
	// TODO: Your code here...
	return
}

// GetSummary implements the AgentImpl interface.
func (s *AgentImpl) GetSummary(ctx context.Context, summaryRequest *agent.GetSummaryRequest) (resp *agent.GetSummaryResponse, err error) {
	// TODO: Your code here...
	return
}

// Ask implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) Ask(ctx context.Context, askRequest *agent.AskRequest) (resp *agent.AskResponse, err error) {
	// TODO: Your code here...
	return
}

// StartSummary implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) StartSummary(ctx context.Context, summaryRequest *agent.StartSummaryRequest) (resp *agent.StartSummaryResponse, err error) {
	// TODO: Your code here...
	return
}

// GetSummary implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) GetSummary(ctx context.Context, summaryRequest *agent.GetSummaryRequest) (resp *agent.GetSummaryResponse, err error) {
	// TODO: Your code here...
	return
}
