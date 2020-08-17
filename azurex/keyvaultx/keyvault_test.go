package keyvaultx_test

import (
	"testing"

	"github.com/SecuritasCrimePrediction/apitools-go/azurex/keyvaultx"
)

func Test_NewKeyvaultFromEnv(t *testing.T) {
	keyvaultx.New("ss")
}
