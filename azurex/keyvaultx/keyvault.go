package keyvaultx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/pkg/errors"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// KeyVault provides an interface towards the Azure Key Vault
// primarily used for upload and download of SSL certificates.
type KeyVault struct {
	keyvault.BaseClient
	vaultURL string
}

func New(kvName string) (KeyVault, error) {
	kv, err := NewFromCLI(kvName)
	if err == nil {
		return kv, err
	}
	return NewFromEnv(kvName)
}

// Create a new connection to an Azure key vault and fetch authentication from environment variables.
// See available env var authentications here
// https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization#use-environment-based-authentication
func NewFromEnv(kvName string) (KeyVault, error) {
	var keyVault KeyVault

	kv := keyvault.New()
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return keyVault, err
	}

	keyVault.BaseClient = kv
	keyVault.Authorizer = authorizer
	keyVault.vaultURL = fmt.Sprintf("https://%s.vault.azure.net/", kvName)

	return keyVault, nil
}

// Create a new connection to an Azure key vault and use authentication credentials from Azure CLI.
// You must be logged in to Azure CLI to use this auth method.
func NewFromCLI(kvName string) (KeyVault, error) {
	var keyVault KeyVault
	kvResourceURL := "https://vault.azure.net"
	token, err := cli.GetTokenFromCLI(kvResourceURL)
	if err != nil {
		return keyVault, err
	}

	adalToken, err := token.ToADALToken()
	if err != nil {
		return keyVault, err
	}
	keyVault.BaseClient = keyvault.New()
	keyVault.Authorizer = autorest.NewBearerAuthorizer(&adalToken)
	keyVault.vaultURL = fmt.Sprintf("https://%s.vault.azure.net/", kvName)

	return keyVault, nil
}

// GetCertificate downloads a certificate and private key from the given Azure key vault.
func (v KeyVault) GetCertificate(ctx context.Context, certName string, secretVersion string, certPassword string) (*x509.Certificate, *rsa.PrivateKey, error) {
	res, err := v.GetSecret(ctx, v.vaultURL, certName, secretVersion)
	if err != nil {
		return nil, nil, err
	}

	expectedContentType := "application/x-pkcs12"
	if len(*res.ContentType) == 0 || *res.ContentType != expectedContentType {
		return nil, nil, fmt.Errorf("invalid secret content type '%v', should be '%v'", *res.ContentType, expectedContentType)
	}

	pfx, err := base64.StdEncoding.DecodeString(*res.Value)
	if err != nil {
		return nil, nil, err
	}

	// Decode pfx to x509.Certificate and rsa.PublicKey
	keyIface, cert, err := pkcs12.Decode(pfx, certPassword)
	if err != nil {
		return nil, nil, errors.Errorf("failed to parse pkcs12: %v", err)
	}
	key, ok := keyIface.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("failed to parse key as rsa.PrivateKey")
	}

	return cert, key, nil
}

// UploadCertificate uploads a new certificate and key pair to the given Azure key vault
func (v KeyVault) UploadCertificate(ctx context.Context, cert *x509.Certificate, key *rsa.PrivateKey, certName string, certPassword string) error {
	// Encode certificate to pkcs12
	pfx, err := pkcs12.Encode(rand.Reader, key, cert, nil, certPassword)
	if err != nil {
		return errors.Errorf("failed to encode pkcs12 cert: %v", err)
	}
	base64Encoded := base64.StdEncoding.EncodeToString(pfx)

	exists, err := v.checkCertExists(ctx, v.vaultURL, certName)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("a certificate with that name already exists")
	}

	// Upload cert
	_, err = v.ImportCertificate(ctx, v.vaultURL, certName, keyvault.CertificateImportParameters{
		Base64EncodedCertificate: &base64Encoded,
		Password:                 &certPassword,
	})

	return err
}

func (v KeyVault) checkCertExists(ctx context.Context, baseURL, certName string) (bool, error) {

	_, _, err := v.GetCertificate(ctx, baseURL, certName, "")
	if err != nil {
		if detailedErr, ok := err.(autorest.DetailedError); ok {
			if detailedErr.StatusCode == 404 {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}
