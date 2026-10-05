package handlers

import (
	"mime"
	"strconv"

	"github.com/gofiber/fiber/v2"
	logs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traces "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (h *ObservabilityHandler) IngestOTLPTraces(c *fiber.Ctx) error {
	orgID, ok := c.Locals("ingestionOrganizationID").(string)
	if !ok || orgID == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body, err := otlpBody(c)
	if err != nil {
		return err
	}
	_, err = h.svc.IngestTraces(c.UserContext(), orgID, body, c.Get("Content-Type"))
	if err != nil {
		return writeOTLPError(c, err)
	}
	return writeOTLPResponse(c, &traces.ExportTraceServiceResponse{})
}

func (h *ObservabilityHandler) IngestOTLPLogs(c *fiber.Ctx) error {
	orgID, ok := c.Locals("ingestionOrganizationID").(string)
	if !ok || orgID == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body, err := otlpBody(c)
	if err != nil {
		return err
	}
	_, err = h.svc.IngestLogs(c.UserContext(), orgID, body, c.Get("Content-Type"))
	if err != nil {
		return writeOTLPError(c, err)
	}
	return writeOTLPResponse(c, &logs.ExportLogsServiceResponse{})
}

func (h *ObservabilityHandler) IngestOTLPMetrics(c *fiber.Ctx) error {
	orgID, ok := c.Locals("ingestionOrganizationID").(string)
	if !ok || orgID == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body, err := otlpBody(c)
	if err != nil {
		return err
	}
	_, err = h.svc.IngestMetrics(c.UserContext(), orgID, body, c.Get("Content-Type"))
	if err != nil {
		return writeOTLPError(c, err)
	}
	return writeOTLPResponse(c, &metrics.ExportMetricsServiceResponse{})
}

func writeOTLPResponse(c *fiber.Ctx, message proto.Message) error {
	typeName, _, err := mime.ParseMediaType(c.Get("Content-Type"))
	if err != nil {
		return c.SendStatus(fiber.StatusUnsupportedMediaType)
	}
	var data []byte
	if typeName == "application/x-protobuf" {
		data, err = proto.Marshal(message)
	} else {
		data, err = protojson.Marshal(message)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Set("Content-Type", typeName)
	return c.Send(data)
}

func (h *ObservabilityHandler) ListMetrics(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "100"))
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	rows, err := h.svc.ListMetrics(c.UserContext(), c.Params("organizationID"), c.Query("name"), limit)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(rows)
}
