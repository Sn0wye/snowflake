package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	logpkg "github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestNewLogAddsServiceField(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "test.log")

	conf := viper.New()
	conf.Set("log.encoding", "json")
	conf.Set("log.level", "info")
	conf.Set("log.log_file_name", logFile)
	conf.Set("env", "test")

	logger := logpkg.NewLog(conf)
	logger.Info("hello", zap.String("foo", "bar"))
	_ = logger.Sync()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if !strings.Contains(string(data), `"service":"helium"`) {
		t.Fatalf("expected log line to contain service field, got: %s", data)
	}
}
