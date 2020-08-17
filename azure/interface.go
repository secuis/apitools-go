package azure

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
)

// Todo: Add update certificate functionality so we don't have to create new certificates as soon as the old expire
type KeyVault interface {
	// GetCertificate downloads a certificate and key from an Azure key vault
	GetCertificate(ctx context.Context, certName string, secretVersion string, certPassword string) (*x509.Certificate, *rsa.PrivateKey, error)

	// UploadCertificate uploads a given certificate and key as certName to an Azure key vault
	UploadCertificate(ctx context.Context, cert *x509.Certificate, key *rsa.PrivateKey, certName string, certPassword string) error
}
