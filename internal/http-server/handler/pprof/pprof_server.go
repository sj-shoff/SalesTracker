package pprof

import (
	"context"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/wb-go/wbf/zlog"

	"sales-tracker/internal/config"
)

type Server struct {
	httpServer      *http.Server
	logger          *zlog.Zerolog
	port            string
	shutdownTimeout time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
}

func NewServer(cfg *config.Config, logger *zlog.Zerolog) *Server {
	return &Server{
		port:            cfg.Pprof.Port,
		shutdownTimeout: cfg.Pprof.ShutdownTimeout,
		readTimeout:     cfg.Pprof.ReadTimeout,
		writeTimeout:    cfg.Pprof.WriteTimeout,
		idleTimeout:     cfg.Pprof.IdleTimeout,
		logger:          logger,
	}
}

func (s *Server) Start() error {
	if s.httpServer != nil {
		return errors.New("pprof server already started")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handlePprof)

	s.httpServer = &http.Server{
		Addr:         ":" + s.port,
		Handler:      mux,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	// s.logger.Info("Pprof server started",
	// 	zappretty.Field("port", s.port),
	// 	zappretty.Field("read_timeout", s.readTimeout.String()),
	// 	zappretty.Field("write_timeout", s.writeTimeout.String()),
	// 	zappretty.Field("idle_timeout", s.idleTimeout.String()),
	// )

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		// s.logger.Debug("pprof server already stopped or not started")
		return nil
	}

	// s.logger.Info("Shutting down pprof server",
	// 	zappretty.Field("timeout", s.shutdownTimeout.String()))

	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}

func (s *Server) handlePprof(w http.ResponseWriter, r *http.Request) {
	if !isLocalhost(r) {
		// s.logger.Warn("Blocked pprof access attempt",
		// 	zappretty.Field("remote_addr", r.RemoteAddr),
		// 	zappretty.Field("path", r.URL.Path),
		// 	zappretty.Field("user_agent", r.UserAgent()),
		// )
		http.Error(w, "Forbidden: access allowed only from localhost", http.StatusForbidden)
		return
	}

	http.DefaultServeMux.ServeHTTP(w, r)
}

func isLocalhost(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	return host == "127.0.0.1" ||
		host == "::1" ||
		host == "localhost" ||
		host == "[::1]"
}
