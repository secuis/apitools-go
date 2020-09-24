package blobx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"gopkg.in/go-playground/validator.v9"
)

type ContainerConfig struct {
	AccountName   string `validate:"required,min=2"`
	AccountKey    string `validate:"required,min=2"`
	ContainerName string `validate:"required,min=2"`
}

type SASOptions struct {
	ValidFrom    time.Time
	ValidTo      time.Time
	ReadAccess   bool
	AddAccess    bool
	CreateAccess bool
	WriteAccess  bool
	DeleteAccess bool
	ListAccess   bool
}

type ContainerConn struct {
	container *storage.Container
}

// Get a connection to a container in an azure storage account.
// Use this func if you already have a connection to an Azure storage account.
func NewContainerConn(storage storage.BlobStorageClient, containerName string) (*ContainerConn, error) {
	container := storage.GetContainerReference(containerName)
	exists, err := container.Exists()
	if err != nil {
		return nil, fmt.Errorf("could not connect to the storage account, err: %v", err)
	}

	if !exists {
		return nil, fmt.Errorf("no container with the name %q exist in this storage account", containerName)
	}

	return &ContainerConn{
		container: container,
	}, nil
}

// Get a connection to a container in an azure storage account.
// Use this func if you don't have a connection to an Azure account already.
func NewAccountContainerConn(conf ContainerConfig) (*ContainerConn, error) {
	v := validator.New()
	if err := v.Struct(conf); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	client, err := storage.NewBasicClient(conf.AccountName, conf.AccountKey)
	if err != nil {
		return nil, fmt.Errorf("could not connect to azure, err: %v", err)
	}

	return NewContainerConn(client.GetBlobService(), conf.ContainerName)
}

func (c *ContainerConn) GetContainerSASURI(ctx context.Context, opts SASOptions) (string, error) {
	return c.container.GetSASURI(storage.ContainerSASOptions{
		ContainerSASPermissions: storage.ContainerSASPermissions{
			BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
				Read:   opts.ReadAccess,
				Add:    opts.AddAccess,
				Create: opts.CreateAccess,
				Write:  opts.WriteAccess,
				Delete: opts.DeleteAccess,
			},
			List: opts.ListAccess,
		},
		OverrideHeaders: storage.OverrideHeaders{},
		SASOptions: storage.SASOptions{
			Start:  opts.ValidFrom,
			Expiry: opts.ValidTo,
		},
	})
}

func (c *ContainerConn) GetBlobSASURI(ctx context.Context, blobName string, opts SASOptions) (string, error) {
	blob := c.container.GetBlobReference(blobName)

	exist, err := blob.Exists()
	if err != nil {
		return "", err
	}

	if !exist {
		return "", fmt.Errorf("blob with the name %q does not exist in this container", blobName)
	}

	return blob.GetSASURI(storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
			Read:   opts.ReadAccess,
			Add:    opts.AddAccess,
			Create: opts.CreateAccess,
			Write:  opts.WriteAccess,
			Delete: opts.DeleteAccess,
		},
		OverrideHeaders: storage.OverrideHeaders{},
		SASOptions: storage.SASOptions{
			Start:  opts.ValidFrom,
			Expiry: opts.ValidTo,
		},
	})
}

func (c *ContainerConn) BlobReader(ctx context.Context, blobName string) (io.ReadCloser, error) {
	blob := c.container.GetBlobReference(blobName)

	exist, err := blob.Exists()
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, fmt.Errorf("blob with the name %q does not exist in this container", blobName)
	}

	return blob.Get(&storage.GetBlobOptions{})
}

func (c *ContainerConn) BlobBytes(ctx context.Context, blob string) ([]byte, error) {
	var err error
	reader, err := c.BlobReader(ctx, blob)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	if err := reader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close blob bytes reader, %v", err)
	}

	return buf.Bytes(), nil
}

func (c *ContainerConn) ListBlobs(ctx context.Context, prefix string) ([]string, error) {
	var blobNames []string

	resp, err := c.container.ListBlobs(storage.ListBlobsParameters{
		Prefix:  prefix,
		Timeout: 15, // seconds
	})
	if err != nil {
		return nil, err
	}

	for _, blob := range resp.Blobs {
		blobNames = append(blobNames, blob.Name)
	}

	return blobNames, nil
}

func (c *ContainerConn) ListBlobsByPattern(ctx context.Context, pattern string) ([]string, error) {
	wildcardParts := strings.Split(pattern, "*")
	dirPrefix := path.Dir(wildcardParts[0])

	// Fix to work with files not in a directory
	if dirPrefix == "." {
		dirPrefix = ""
	}

	blobNames, err := c.ListBlobs(ctx, dirPrefix)
	if err != nil {
		return nil, err
	}

	var matchingBlobNames []string
	for _, blobName := range blobNames {
		matched, err := path.Match(pattern, blobName)
		if err != nil {
			return nil, fmt.Errorf("unexpected error when matching patterns: %w", err)
		}
		if matched {
			matchingBlobNames = append(matchingBlobNames, blobName)
		}
	}

	return matchingBlobNames, nil
}

func (c *ContainerConn) TruncateBlob(ctx context.Context, reader io.Reader, blobName string) error {
	blob := c.container.GetBlobReference(blobName)
	if _, err := blob.DeleteIfExists(&storage.DeleteBlobOptions{
		Timeout: 15,
	}); err != nil {
		return err
	}

	return c.AppendBlob(ctx, reader, blobName)
}

func (c *ContainerConn) AppendBlob(ctx context.Context, reader io.Reader, blobName string) error {
	blob := c.container.GetBlobReference(blobName)

	exist, err := blob.Exists()
	if err != nil {
		return err
	}

	if !exist {
		if err := blob.PutAppendBlob(&storage.PutBlobOptions{}); err != nil {
			return err
		}
	}

	buf := make([]byte, 1*1024*1024)
	for {
		_, err := reader.Read(buf)
		if err == io.EOF {
			break
		}

		if err := blob.AppendBlock(buf, &storage.AppendBlockOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// blobname is the name of the blob to create a lockfile for
// if lockfile already exist LockfileAlreadyExist error will be returned
func (c *ContainerConn) CreateLockFile(ctx context.Context, blobName string) error {
	lockBlobName := fmt.Sprintf("%s.%s", blobName, ".LOCK")
	blob := c.container.GetBlobReference(lockBlobName)

	exist, err := blob.Exists()
	if err != nil {
		return err
	}

	if exist {
		return LockfileAlreadyExist
	}

	if err := blob.PutAppendBlob(&storage.PutBlobOptions{
		Timeout: 15,
	}); err != nil {
		return err
	}

	return nil
}

// blobname is the name of the blob to delete a lockfile for
// will not give an error if the lockfile does not exist
func (c *ContainerConn) DeleteLockFile(ctx context.Context, blobName string) error {
	lockBlobName := fmt.Sprintf("%s.%s", blobName, ".LOCK")
	blob := c.container.GetBlobReference(lockBlobName)

	if _, err := blob.DeleteIfExists(&storage.DeleteBlobOptions{
		Timeout: 15,
	}); err != nil {
		return err
	}

	return nil
}