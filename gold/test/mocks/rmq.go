package mocks

import (
	"sync"

	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"go.uber.org/zap"
)

type MockRmq struct {
	mu       sync.Mutex
	messages []string
}

func NewMockRmq() *MockRmq {
	return &MockRmq{messages: make([]string, 0)}
}

func (m *MockRmq) Produce(queueName, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockRmq) ProduceWithHeaders(queueName, message, correlationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockRmq) Messages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.messages))
	copy(cp, m.messages)
	return cp
}

func TestLogger() *logger.Logger {
	zapLogger := zap.NewNop()
	return &logger.Logger{Logger: zapLogger}
}
