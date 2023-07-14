package rpc

import (
	"bytes"
	"context"
	"io"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type IPDF2ImagesService interface {
	ConvertPDF2Images(ctx context.Context, containerName string, blobFolderPath string, fileName string, buffer bytes.Buffer) error
}

type RPCHandler struct {
	UnimplementedPDF2ImagesServiceServer
	service IPDF2ImagesService
	logger  *otelzap.SugaredLogger
}

func NewRPCHandler(service IPDF2ImagesService, logger *otelzap.SugaredLogger) *RPCHandler {
	return &RPCHandler{service: service, logger: logger}
}

func (h *RPCHandler) ConvertPDF2Images(stream PDF2ImagesService_ConvertPDF2ImagesServer) error {
	ctx, err := tracer.AddTraceIDToContext(stream.Context())
	if err != nil {
		return ErrorLogger(ctx, h.logger, err)
	}

	ctx, span := tracer.NewSpan(ctx, "RPCHandler.ConvertPDF2Images", nil)
	defer span.End()

	req, err := stream.Recv()
	if err != nil {
		return ErrorLogger(ctx, h.logger, err)
	}

	meta := req.GetMeta()
	containerName := meta.GetContainerName()
	blobFolderPath := meta.GetBlobFolderPath()
	fileName := meta.GetFileName()

	buffer := bytes.Buffer{}
	for {
		req, err = stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ErrorLogger(ctx, h.logger, err)
		}

		chunk := req.GetChunk()
		if _, err = buffer.Write(chunk); err != nil {
			return ErrorLogger(ctx, h.logger, err)
		}
	}

	if span.IsRecording() {
		tracer.AddSpanEvents(span, "ConvertPDF2Images input", map[string]string{
			"containerName":  containerName,
			"blobFolderPath": blobFolderPath,
			"fileName":       fileName,
			"file":           buffer.String(),
		})
	}

	if err = h.service.ConvertPDF2Images(ctx, containerName, blobFolderPath, fileName, buffer); err != nil {
		return ErrorLogger(ctx, h.logger, err)
	}

	err = stream.SendAndClose(&EmptyResponse{})
	if err != nil {
		return ErrorLogger(ctx, h.logger, err)
	}

	return nil
}
