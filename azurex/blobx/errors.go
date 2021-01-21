package blobx

import (
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/pkg/errors"
)

var (
	ErrBlobNotFound          = errors.New("no blob with that name found")
	ErrBlobBusy              = errors.New("blob is busy, can't perform this action")
	ErrUnknownStorageAccount = errors.New("unknown storage account")
	ErrUploadFailed          = errors.New("blob upload failed")
	LockfileAlreadyExist     = errors.New("lockfile already exist")
)

// todo: add more errors
func ParseAzureError(err error) error {
	if err != nil {
		azErr, ok := err.(storage.AzureStorageServiceError)
		if !ok {
			return err
		}
		switch azErr.StatusCode {
		case 404:
			return ErrBlobNotFound
		case 409:
			return ErrBlobBusy
		}
	}

	return err
}
