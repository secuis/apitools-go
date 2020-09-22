package blobx

import (
	"context"
	"fmt"
	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
	"io"
	"sync"

	"github.com/Azure/azure-sdk-for-go/storage"
	"gopkg.in/go-playground/validator.v9"
)

type AccountConfig struct {
	Name string `validate:"required,min=2"`
	Key  string `validate:"required,min=2"`
}

type AccountConn struct {
	name         string
	containers   map[string]*ContainerConn
	pipe         pipeline.Pipeline
	blobService  storage.BlobStorageClient
	containerMtx sync.RWMutex
}

// NewAccount returns an Azure implementation of a Storage account
func NewAccount(config *AccountConfig) (*AccountConn, error) {
	v := validator.New()
	if err := v.Struct(config); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	client, err := storage.NewBasicClient(config.Name, config.Key)
	if err != nil {
		return nil, fmt.Errorf("could not connect to azure, err: %v", err)
	}

	// pipeline is needed for uploading big files
	credential, err := azblob.NewSharedKeyCredential(config.Name, config.Key)
	if err != nil {
		return nil, errors.Errorf("Invalid credentials with error: " + err.Error())
	}

	pipe := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	return &AccountConn{
		containers:  map[string]*ContainerConn{},
		name:        config.Name,
		blobService: client.GetBlobService(),
		pipe: pipe,
	}, nil
}

// Get the connection to a container in the storage account.
func (a *AccountConn) NewContainer(containerName string) (*ContainerConn, error) {
	a.containerMtx.Lock()
	defer a.containerMtx.Unlock()

	containerConn, err := NewContainerConn(a.blobService, containerName, a.pipe)
	if err != nil {
		return nil, err
	}

	a.containers[containerName] = containerConn
	return containerConn, nil
}

func (a *AccountConn) GetContainerSASURI(ctx context.Context, container string, opts SASOptions) (string, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return "", err
		}
	}

	return containerConn.GetContainerSASURI(ctx, opts)
}

func (a *AccountConn) GetBlobSASURI(ctx context.Context, container string, blobName string, opts SASOptions) (string, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return "", err
		}
	}

	return containerConn.GetBlobSASURI(ctx, blobName, opts)
}

func (a *AccountConn) BlobReader(ctx context.Context, container string, blob string) (io.ReadCloser, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return nil, err
		}
	}

	return containerConn.BlobReader(ctx, blob)
}

func (a *AccountConn) BlobBytes(ctx context.Context, container string, blob string) ([]byte, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return nil, err
		}
	}

	return containerConn.BlobBytes(ctx, blob)
}

func (a *AccountConn) ListBlobs(ctx context.Context, container string, prefix string) ([]string, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return nil, err
		}
	}

	return containerConn.ListBlobs(ctx, prefix)
}

func (a *AccountConn) ListBlobsByPattern(ctx context.Context, container string, pattern string) ([]string, error) {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return nil, err
		}
	}

	return containerConn.ListBlobsByPattern(ctx, pattern)
}

func (a *AccountConn) TruncateBlob(ctx context.Context, container string, reader io.Reader, blobName string) error {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return err
		}
	}
	return containerConn.TruncateBlob(ctx, reader, blobName)
}

func (a *AccountConn) AppendBlob(ctx context.Context, container string, reader io.Reader, blobName string) error {
	containerConn, exist := a.containers[container]
	if !exist {
		var err error
		containerConn, err = a.NewContainer(container)
		if err != nil {
			return err
		}
	}
	return containerConn.AppendBlob(ctx, reader, blobName)
}