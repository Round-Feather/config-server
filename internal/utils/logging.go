package utils

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func GetLogFields(c echo.Context) log.Fields {
	ctx := c.Request().Context()
	span := trace.SpanFromContext(ctx)

	traceId := span.SpanContext().TraceID().String()
	spanId := span.SpanContext().SpanID().String()

	return log.Fields{
		"spanId":     spanId,
		"traceId":    traceId,
		"commitHash": c.Get("hash"),
	}
}
