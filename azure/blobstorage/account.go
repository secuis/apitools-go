package blobstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

type AccountConfig struct {
	Name string `validate:"required,min=2"`
	Key  string `validate:"required,min=2"`
}

type StorageAccount interface {
	// BlobReader returns an io.ReadCloser which can be used to read the content of a blob
	BlobReader(ctx context.Context, container string, blob string) (io.ReadCloser, error)

	// BlobBytes returns the bytes of a blob
	BlobBytes(ctx context.Context, container string, blob string) ([]byte, error)

	// ListBlobs returns a list of blob names with the provided prefix that was found in the provided container
	ListBlobs(ctx context.Context, container string, prefix string) ([]string, error)

	// ListBlobsByPattern returns a list of blob names given the provided pattern.
	// The provided pattern is used to match blob names using the path.Match function
	ListBlobsByPattern(ctx context.Context, container string, pattern string) ([]string, error)

	// UploadBlob takes a reader of a blob to be stream-uploaded to an azure blob storage
	UploadBlob(ctx context.Context, container string, reader io.Reader, blobName string) error
}

type storageAccount struct {
	// Pipelines are threadsafe and may be shared
	pipeline      pipeline.Pipeline
	name          string
	containerURLs map[string]*azblob.ContainerURL
}

// NewAccount returns an Azure implementation of a StorageAccount
func NewAccount(config *AccountConfig) (StorageAccount, error) {
	v := validator.New()
	if err := v.Struct(config); err != nil {
		return nil, errors.Errorf("Config error: " + err.Error())
	}

	a := &storageAccount{
		containerURLs: map[string]*azblob.ContainerURL{},
		name:          config.Name,
	}

	credential, err := azblob.NewSharedKeyCredential(config.Name, config.Key)
	if err != nil {
		return nil, errors.Errorf("Invalid credentials with error: " + err.Error())
	}

	a.pipeline = azblob.NewPipeline(credential, azblob.PipelineOptions{})

	return a, nil
}

func (sa *storageAccount) newContainer(containerName string) *azblob.ContainerURL {
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", sa.name, containerName))
	newURL := azblob.NewContainerURL(*URL, sa.pipeline)
	return &newURL
}

func (sa *storageAccount) containerURL(containerName string) (*azblob.ContainerURL, error) {
	if _, ok := sa.containerURLs[containerName]; ok {
		return sa.containerURLs[containerName], nil
	}

	sa.containerURLs[containerName] = sa.newContainer(containerName)

	if sa.containerURLs[containerName] == nil {
		return nil, errors.Errorf("Could not create container url for container %s", containerName)
	}

	return sa.containerURLs[containerName], nil
}

func (sa storageAccount) BlobReader(ctx context.Context, container string, blob string) (io.ReadCloser, error) {
	cURL, err := sa.containerURL(container)
	if err != nil {
		return nil, err
	}

	bURL := cURL.NewBlobURL(blob)
	resp, err := bURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}

	return resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20}), nil
}

func (sa storageAccount) BlobBytes(ctx context.Context, container string, blob string) ([]byte, error) {
	reader, err := sa.BlobReader(ctx, container, blob)

	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	_ = reader.Close()

	return buf.Bytes(), nil
}

func (sa storageAccount) ListBlobs(ctx context.Context, container string, prefix string) ([]string, error) {
	var blobNames []string

	cURL, err := sa.containerURL(container)
	if err != nil {
		return nil, err
	}

	response, err := cURL.ListBlobsFlatSegment(ctx, azblob.Marker{}, azblob.ListBlobsSegmentOptions{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	for _, blob := range response.Segment.BlobItems {
		blobNames = append(blobNames, blob.Name)
	}

	return blobNames, nil
}

func (sa storageAccount) ListBlobsByPattern(ctx context.Context, container string, pattern string) ([]string, error) {
	wildcardParts := strings.Split(pattern, "*")
	dirPrefix := path.Dir(wildcardParts[0])

	// Fix to work with files not in a directory
	if dirPrefix == "." {
		dirPrefix = ""
	}

	blobNames, err := sa.ListBlobs(ctx, container, dirPrefix)
	if err != nil {
		return nil, err
	}

	var matchingBlobNames []string
	for _, blobName := range blobNames {
		matched, err := path.Match(pattern, blobName)
		if err != nil {
			return nil, errors.Errorf("unexpected error when matching patterns: %w", err)
		}
		if matched {
			matchingBlobNames = append(matchingBlobNames, blobName)
		}
	}

	return matchingBlobNames, nil
}

func (sa storageAccount) UploadBlob(ctx context.Context, container string, reader io.Reader, blobName string) error {
	cURL, err := sa.containerURL(container)
	if err != nil {
		return err
	}

	bURL := cURL.NewBlockBlobURL(blobName)

	resp, err := azblob.UploadStreamToBlockBlob(ctx, reader, bURL, azblob.UploadStreamToBlockBlobOptions{
		BufferSize: 1 * 1024 * 1024,
		MaxBuffers: 3,
	})

	if err != nil {
		return err
	}

	if resp.Response().StatusCode != 201 {
		return ErrUploadFailed
	}

	return nil
}
