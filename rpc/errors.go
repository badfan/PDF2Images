package rpc

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

func ErrorLogger(ctx context.Context, logger *otelzap.SugaredLogger, err error) error {
	logger.Ctx(ctx).Warnw("response to an error request", "response_err_msg", err.Error())
	return err
}
