package server

import (
	"database/sql"
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type ConsulAgent interface {
	Stop()
}

type consulAgent struct {
	logger  *zap.Logger
	db      *sql.DB
	config  Config
	agent   *consul.Agent
	TTL     time.Duration
	stopped *atomic.Bool
}

func StartConsulAgent(logger *zap.Logger, startupLogger *zap.Logger, db *sql.DB, config Config, serverVersion string) ConsulAgent {
	s := &consulAgent{
		logger:  logger,
		db:      db,
		config:  config,
		stopped: atomic.NewBool(false),
		TTL:     time.Duration(config.GetConsul().TTLms * int(time.Millisecond)),
	}
	startupLogger.Info("Initiate Consul agent registration",
		zap.String("address", config.GetConsul().Address),
		zap.Int("port", config.GetConsul().Port),
		zap.Duration("ttl", s.TTL))
	cfg := consul.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", config.GetConsul().Address, config.GetConsul().Port)
	if c, err := consul.NewClient(cfg); err != nil {
		startupLogger.Info("Consul agent registration disabled", zap.Error(err))
	} else {
		s.agent = c.Agent()
		if err := s.agent.ServiceRegister(&consul.AgentServiceRegistration{
			Kind: consul.ServiceKindAPIGateway,
			Name: config.GetName(),
			Tags: []string{serverVersion},
			Check: &consul.AgentServiceCheck{
				TTL: s.TTL.String(),
			},
		}); err != nil {
			startupLogger.Info("Consul agent registration disabled")
		} else {
			startupLogger.Info("Starting Consul server health check")
			go s.updateTTL(func() error {
				if err := db.Ping(); err != nil {
					return err
				}
				return nil
			})
		}
	}
	return s
}

func (s *consulAgent) updateTTL(check func() error) {
	ticker := time.NewTicker(s.TTL / 2)
	for range ticker.C {
		if s.stopped.Load() {
			s.agent.ServiceDeregister(s.config.GetName())
			break
		} else if err := check(); err != nil {
			if agentErr := s.agent.FailTTL(
				"service:"+s.config.GetName(), err.Error()); agentErr != nil {
				s.logger.Warn("Consul agent failed:", zap.Error(agentErr))
			}
		} else {
			if agentErr := s.agent.PassTTL(
				"service:"+s.config.GetName(), ""); agentErr != nil {
				s.logger.Warn("Consul agent passed:", zap.Error(agentErr))
			}
		}
	}
}

func (s *consulAgent) Stop() {
	s.stopped.Store(false)
}
