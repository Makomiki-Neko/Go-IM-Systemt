package svc

import "IMM/rpc/ai/internal/config"

type ServiceContext struct {
	Config config.Config
	LLM    *OpenAIClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Create LLM Client
	llm := NewOpenAIClient(c.LLM.Key, c.LLM.Api)
	return &ServiceContext{
		Config: c,
		LLM:    llm,
	}
}
