package storage

import (
	"context"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type BlobFileStorageClient struct {
	client blob.IBlobClient
	logger *otelzap.SugaredLogger
}

func NewBlobFileStorageClient(logger *otelzap.SugaredLogger) (*BlobFileStorageClient, error) {
	client, err := blob.NewBlob(
		blob.BlobConfig{
			AccountName: viper.GetString("blob_account_name"),
			AccountKey:  viper.GetString("blob_account_key"),
		},
		time.Duration(viper.GetInt64("blob_timeout"))*time.Second)
	if err != nil {
		return nil, err
	}

	return &BlobFileStorageClient{client: client, logger: logger}, nil
}

func (c *BlobFileStorageClient) UploadFolderToBlobStorage(ctx context.Context, containerName string, blobFolderPath string, tempImagesFolder string) error {
	ctx, span := tracer.NewSpan(ctx, "FileStorageClient.UploadFolderToBlobStorage", nil)
	defer span.End()

	defer os.RemoveAll(tempImagesFolder)

	if _, err := c.client.UploadFolder(ctx, containerName, blobFolderPath, tempImagesFolder); err != nil {
		return err
	}

	return nil
}
