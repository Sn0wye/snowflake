package mocks

import (
	"sync"

	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"go.uber.org/zap"
)

type MockRmq struct {
	mu             sync.Mutex
	messages       []string
	correlationIDs []string
}

func NewMockRmq() *MockRmq {
	return &MockRmq{
		messages:       make([]string, 0),
		correlationIDs: make([]string, 0),
	}
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
	m.correlationIDs = append(m.correlationIDs, correlationID)
	return nil
}

func (m *MockRmq) Messages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.messages))
	copy(cp, m.messages)
	return cp
}

func (m *MockRmq) CorrelationIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.correlationIDs))
	copy(cp, m.correlationIDs)
	return cp
}

func TestLogger() *logger.Logger {
	zapLogger := zap.NewNop()
	return &logger.Logger{Logger: zapLogger}
}
