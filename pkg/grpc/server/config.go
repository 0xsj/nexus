package server

import (
	"time"

	"github.com/0xsj/nexus/pkg/config"
)

// Config holds gRPC server configuration
type Config struct {
	Port                  int           `env:"GRPC_PORT" default:"9090"`
	MaxConnectionIdle     time.Duration `env:"GRPC_MAX_CONNECTION_IDLE" default:"5m"`
	MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" default:"2h"`
	MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" default:"5m"`
	KeepAliveTime         time.Duration `env:"GRPC_KEEPALIVE_TIME" default:"2h"`
	KeepAliveTimeout      time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT" default:"20s"`
	MaxRecvMsgSize        int           `env:"GRPC_MAX_RECV_MSG_SIZE" default:"4194304"` // 4MB
	MaxSendMsgSize        int           `env:"GRPC_MAX_SEND_MSG_SIZE" default:"4194304"` // 4MB
	EnableReflection      bool          `env:"GRPC_ENABLE_REFLECTION" default:"true"`
}

// Load loads server configuration from environment
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	return config.Load(cfg)
}
