package mqtls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadClientTLSConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	certPEM, keyPEM := generateSelfSignedCertPEM(t)
	certFile := filepath.Join(dir, "client.crt")
	keyFile := filepath.Join(dir, "client.key")
	caFile := filepath.Join(dir, "ca.crt")

	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	if err := os.WriteFile(caFile, certPEM, 0o600); err != nil {
		t.Fatalf("write ca: %v", err)
	}

	cfg, err := LoadClientTLSConfig(certFile, keyFile, caFile, "")
	if err != nil {
		t.Fatalf("LoadClientTLSConfig: %v", err)
	}
	if cfg == nil || len(cfg.Certificates) != 1 {
		t.Fatalf("unexpected tls config: %#v", cfg)
	}
}

func TestLoadClientTLSConfigFromEnvMissingPassword(t *testing.T) {
	dir := t.TempDir()
	certPEM, keyPEM := generateEncryptedKeyPEM(t, "secret")
	certFile := filepath.Join(dir, "client.crt")
	keyFile := filepath.Join(dir, "client.key")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	_, err := LoadClientTLSConfigFromEnv(certFile, keyFile, "", "TLS_KEY_PASSWORD")
	if err == nil {
		t.Fatal("expected error when password env is missing")
	}
}

func TestParsePrivateKeyEncryptedPEM(t *testing.T) {
	_, keyPEM := generateEncryptedKeyPEM(t, "secret")

	if _, err := parsePrivateKey(keyPEM, ""); err == nil {
		t.Fatal("expected password-required error")
	}
	if _, err := parsePrivateKey(keyPEM, "secret"); err != nil {
		t.Fatalf("expected key parse to succeed with password: %v", err)
	}
}

func generateSelfSignedCertPEM(t *testing.T) ([]byte, []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPEM, keyPEM
}

func generateEncryptedKeyPEM(t *testing.T, password string) ([]byte, []byte) {
	t.Helper()
	certPEM, keyPEM := generateSelfSignedCertPEM(t)

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to decode key PEM")
	}
	encrypted, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(password), x509.PEMCipherAES256)
	if err != nil {
		t.Fatalf("encrypt key: %v", err)
	}
	return certPEM, pem.EncodeToMemory(encrypted)
}
