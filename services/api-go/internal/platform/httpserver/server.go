package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	Handler         http.Handler
	Logger          *zerolog.Logger
}

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	logger          zerolog.Logger
}

func New(cfg Config) *Server {
	logger := zerolog.Nop()
	if cfg.Logger != nil {
		logger = *cfg.Logger
	}

	handler := cfg.Handler
	if handler == nil {
		handler = http.NewServeMux()
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
		shutdownTimeout: cfg.ShutdownTimeout,
		logger:          logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info().Str("addr", s.httpServer.Addr).Msg("api listening")
		err := s.httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	}
}
