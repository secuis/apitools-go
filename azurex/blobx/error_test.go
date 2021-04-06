package blobx

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"testing"
)

func Test_ErrorParsing(t *testing.T) {
	err := ParseAzureError(storage.AzureStorageServiceError{
		Code:       "a code",
		Message:    "random error",
		StatusCode: 400,
	})
	if err == nil {
		t.Errorf("error parsing failed, it cannot handle azure errors")
	}

	err = ParseAzureError(fmt.Errorf("random err"))
	if err == nil {
		t.Errorf("error parsing failed, it cannot handle NON-azure errors")
	}
}
