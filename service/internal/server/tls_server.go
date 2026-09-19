package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TLSServerConfig TLS服务器配置
type TLSServerConfig struct {
	HTTPPort       int
	HTTPSPort      int
	CertFile       string
	KeyFile        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout   time.Duration
	Logger        *zap.Logger
}

// TLSServer TLS服务器
type TLSServer struct {
	config       *TLSServerConfig
	httpServer   *http.Server
	httpsServer  *http.Server
	handler      http.Handler
	logger       *zap.Logger
}

// NewTLSServer 创建TLS服务器
func NewTLSServer(handler http.Handler, config *TLSServerConfig) *TLSServer {
	return &TLSServer{
		config:  config,
		handler: handler,
		logger:  config.Logger,
	}
}

// Start 启动HTTP和HTTPS服务器
func (s *TLSServer) Start() error {
	// 配置HTTP服务器（自动重定向到HTTPS）
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler:      s.createRedirectHandler(),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// 配置HTTPS服务器
	s.httpsServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.HTTPSPort),
		Handler:      s.handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
		},
	}

	// 启动HTTP服务器
	go func() {
		s.logger.Info("HTTP server started (redirecting to HTTPS)", 
			zap.Int("port", s.config.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// 启动HTTPS服务器
	s.logger.Info("HTTPS server started", 
		zap.Int("port", s.config.HTTPSPort),
		zap.String("cert", s.config.CertFile),
		zap.String("key", s.config.KeyFile))
	
	if err := s.httpsServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTPS server error: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (s *TLSServer) Stop() error {
	var errors []error

	if s.httpServer != nil {
		if err := s.httpServer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server close error: %w", err))
		}
	}

	if s.httpsServer != nil {
		if err := s.httpsServer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("HTTPS server close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errors)
	}

	return nil
}

// createRedirectHandler 创建HTTP重定向到HTTPS的处理器
func (s *TLSServer) createRedirectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		targetURL := fmt.Sprintf("https://%s%s", host, r.URL.RequestURI())
		
		// 添加HSTS头
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
	})
}

// GetHTTPServer 获取HTTP服务器
func (s *TLSServer) GetHTTPServer() *http.Server {
	return s.httpServer
}

// GetHTTPSServer 获取HTTPS服务器
func (s *TLSServer) GetHTTPSServer() *http.Server {
	return s.httpsServer
}
