package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TLSConfig holds mTLS configuration for gRPC clients
type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// LoadClientTLS loads mTLS credentials for gRPC clients
func LoadClientTLS(cfg TLSConfig) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

// GetDialOptions returns gRPC dial options with mTLS if configured
func GetDialOptions(cfg TLSConfig) []grpc.DialOption {
	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CAFile != "" {
		creds, err := LoadClientTLS(cfg)
		if err == nil {
			return []grpc.DialOption{grpc.WithTransportCredentials(creds)}
		}
	}
	// Fallback to insecure for development
	return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
}
