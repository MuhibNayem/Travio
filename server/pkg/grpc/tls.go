package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// LoadServerTLS creates TLS credentials for a gRPC server with mTLS
// certFile: Server certificate
// keyFile: Server private key
// caFile: CA certificate for client verification (mTLS)
func LoadServerTLS(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert: %w", err)
	}

	// Load CA cert for client verification
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // mTLS: require client cert
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

// LoadClientTLS creates TLS credentials for a gRPC client with mTLS
// certFile: Client certificate
// keyFile: Client private key
// caFile: CA certificate for server verification
func LoadClientTLS(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	// Load CA cert for server verification
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

// LoadServerTLSOptional loads TLS if cert files exist, otherwise returns nil (for development)
func LoadServerTLSOptional(certFile, keyFile, caFile string) credentials.TransportCredentials {
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil
	}
	creds, err := LoadServerTLS(certFile, keyFile, caFile)
	if err != nil {
		return nil
	}
	return creds
}

// LoadClientTLSOptional loads TLS if cert files exist, otherwise returns nil (for development)
func LoadClientTLSOptional(certFile, keyFile, caFile string) credentials.TransportCredentials {
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil
	}
	creds, err := LoadClientTLS(certFile, keyFile, caFile)
	if err != nil {
		return nil
	}
	return creds
}
