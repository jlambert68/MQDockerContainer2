package mqtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

// LoadClientTLSConfig loads a client certificate/key (optionally encrypted) and an optional CA bundle.
// If keyPassword is empty, unencrypted keys are assumed.
func LoadClientTLSConfig(certFile, keyFile, caFile, keyPassword string) (*tls.Config, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("read cert: %w", err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	cert, err := parseClientCertificate(certPEM, keyPEM, keyPassword)
	if err != nil {
		return nil, err
	}

	var roots *x509.CertPool
	if caFile != "" {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read ca: %w", err)
		}
		roots = x509.NewCertPool()
		if !roots.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse ca: no certificates found")
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// LoadClientTLSConfigFromEnv loads the key password from the given environment variable.
func LoadClientTLSConfigFromEnv(certFile, keyFile, caFile, passwordEnv string) (*tls.Config, error) {
	if passwordEnv == "" {
		return LoadClientTLSConfig(certFile, keyFile, caFile, "")
	}
	pass := os.Getenv(passwordEnv)
	if pass == "" {
		return nil, fmt.Errorf("missing key password env: %s", passwordEnv)
	}
	return LoadClientTLSConfig(certFile, keyFile, caFile, pass)
}

// LoadClientTLSConfigFromP12 loads a PKCS#12 (.p12/.pfx) bundle.
// The bundle can include the client cert, key, and optional CA chain.
func LoadClientTLSConfigFromP12(p12File, caFile, password string) (*tls.Config, error) {
	p12Data, err := os.ReadFile(p12File)
	if err != nil {
		return nil, fmt.Errorf("read p12: %w", err)
	}

	key, cert, err := pkcs12.Decode(p12Data, password)
	if err != nil {
		return nil, fmt.Errorf("decode p12: %w", err)
	}

	tlsCert := tls.Certificate{
		PrivateKey:  key,
		Certificate: [][]byte{cert.Raw},
		Leaf:        cert,
	}

	roots, err := loadRootCAs(caFile, nil)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		RootCAs:      roots,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// LoadClientTLSConfigFromP12Env loads a PKCS#12 bundle using the password from an env var.
func LoadClientTLSConfigFromP12Env(p12File, caFile, passwordEnv string) (*tls.Config, error) {
	if passwordEnv == "" {
		return LoadClientTLSConfigFromP12(p12File, caFile, "")
	}
	pass := os.Getenv(passwordEnv)
	if pass == "" {
		return nil, fmt.Errorf("missing p12 password env: %s", passwordEnv)
	}
	return LoadClientTLSConfigFromP12(p12File, caFile, pass)
}

func parseClientCertificate(certPEM, keyPEM []byte, keyPassword string) (tls.Certificate, error) {
	var certDERs [][]byte
	rest := certPEM
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			certDERs = append(certDERs, block.Bytes)
		}
	}
	if len(certDERs) == 0 {
		return tls.Certificate{}, fmt.Errorf("parse cert: no certificates found")
	}

	key, err := parsePrivateKey(keyPEM, keyPassword)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: certDERs,
		PrivateKey:  key,
	}, nil
}

func parsePrivateKey(keyPEM []byte, keyPassword string) (any, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("parse key: no PEM block found")
	}

	// Encrypted PKCS#8
	if block.Type == "ENCRYPTED PRIVATE KEY" {
		return nil, fmt.Errorf("parse key: encrypted PKCS#8 not supported; use PKCS#12 or convert the key")
	}

	// PEM encryption headers (legacy PEM encryption)
	if x509.IsEncryptedPEMBlock(block) {
		if keyPassword == "" {
			return nil, fmt.Errorf("parse key: password required for encrypted key")
		}
		der, err := x509.DecryptPEMBlock(block, []byte(keyPassword))
		if err != nil {
			return nil, fmt.Errorf("decrypt key: %w", err)
		}
		return parsePrivateKeyDER(der, block.Type)
	}

	return parsePrivateKeyDER(block.Bytes, block.Type)
}

func parsePrivateKeyDER(der []byte, blockType string) (any, error) {
	switch blockType {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(der)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(der)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(der)
	default:
		// Try PKCS#8 as a fallback for unknown labels.
		key, err := x509.ParsePKCS8PrivateKey(der)
		if err != nil {
			return nil, fmt.Errorf("parse key: unsupported key type %q", blockType)
		}
		return key, nil
	}
}

func loadRootCAs(caFile string, extraCerts []*x509.Certificate) (*x509.CertPool, error) {
	var roots *x509.CertPool
	if caFile != "" {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read ca: %w", err)
		}
		roots = x509.NewCertPool()
		if !roots.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse ca: no certificates found")
		}
	}

	if len(extraCerts) > 0 {
		if roots == nil {
			roots = x509.NewCertPool()
		}
		for _, c := range extraCerts {
			roots.AddCert(c)
		}
	}

	return roots, nil
}
