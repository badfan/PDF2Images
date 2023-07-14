package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

var (
	ErrPDFNotConverted = errors.New(".pdf file not converted")
	ErrPDFNotUploaded  = errors.New(".pdf file was converted but not uploaded")
)

type IBlobFileStorageClient interface {
	UploadFolderToBlobStorage(ctx context.Context, containerName string, blobFolderPath string, tempImagesFolder string) error
}

type PDF2ImagesService struct {
	blobFileStorageClient IBlobFileStorageClient
	logger                *otelzap.SugaredLogger
}

func NewPDF2ImagesService(blobFileStorageClient IBlobFileStorageClient, logger *otelzap.SugaredLogger) *PDF2ImagesService {
	return &PDF2ImagesService{blobFileStorageClient: blobFileStorageClient, logger: logger}
}

func (s *PDF2ImagesService) ConvertPDF2Images(ctx context.Context, containerName string, blobFolderPath string, fileName string, buffer bytes.Buffer) error {
	ctx, span := tracer.NewSpan(ctx, "PDF2ImagesService.ConvertPDF2Images", nil)
	defer span.End()

	tempImagesFolder, err := convert(fileName, buffer)
	if err != nil {
		s.logger.Ctx(ctx).Errorw("error occurred while converting .pdf file", "error", err.Error())
		return ErrPDFNotConverted
	}

	if err = s.blobFileStorageClient.UploadFolderToBlobStorage(ctx, containerName, blobFolderPath, tempImagesFolder); err != nil {
		s.logger.Ctx(ctx).Errorw("error occurred while uploading images", "error", err.Error())
		return ErrPDFNotUploaded
	}

	return nil
}

func convert(fileName string, buffer bytes.Buffer) (string, error) {
	file, err := os.CreateTemp("", fileName)
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())

	if _, err = file.Write(buffer.Bytes()); err != nil {
		return "", err
	}

	doc, err := fitz.New(file.Name())
	if err != nil {
		return "", err
	}
	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "tmpImages")
	if err != nil {
		panic(err)
	}

	// Extract pages as images
	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			return "", err
		}

		namePattern := fmt.Sprintf("*-%s(%03d).jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)), n)
		f, err := os.CreateTemp(tmpDir, namePattern)
		if err != nil {
			return "", err
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			return "", err
		}

		f.Close()
	}

	return tmpDir, nil
}
