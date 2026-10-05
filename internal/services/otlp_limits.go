package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const maxExpandedTelemetryBytes = 16 * 1024 * 1024

type otlpParseBudget struct {
	fields   int
	messages int
	records  int
	metrics  int
}

func preflightProtobuf(raw []byte, descriptor protoreflect.MessageDescriptor, budget *otlpParseBudget, depth int) error {
	if depth > 64 {
		return fmt.Errorf("OTLP nesting exceeds limit")
	}
	budget.messages++
	if budget.messages > 100000 {
		return fmt.Errorf("OTLP message allocation limit exceeded")
	}
	switch descriptor.Name() {
	case "Span", "LogRecord", "NumberDataPoint", "HistogramDataPoint", "ExponentialHistogramDataPoint", "SummaryDataPoint":
		budget.records++
	case "Metric":
		budget.metrics++
	}
	if budget.records > maxOTLPRecords || budget.metrics > maxOTLPRecords {
		return fmt.Errorf("OTLP record limit exceeded")
	}
	for len(raw) > 0 {
		budget.fields++
		if budget.fields > 300000 {
			return fmt.Errorf("OTLP field allocation limit exceeded")
		}
		number, wireType, tagSize := protowire.ConsumeTag(raw)
		if tagSize < 0 {
			return fmt.Errorf("invalid OTLP protobuf tag")
		}
		if wireType == protowire.StartGroupType || wireType == protowire.EndGroupType {
			return fmt.Errorf("OTLP protobuf groups are unsupported")
		}
		raw = raw[tagSize:]
		fieldSize := protowire.ConsumeFieldValue(number, wireType, raw)
		if fieldSize < 0 {
			return fmt.Errorf("invalid OTLP protobuf field")
		}
		field := descriptor.Fields().ByNumber(protoreflect.FieldNumber(number))
		if wireType == protowire.BytesType && field != nil && field.Kind() == protoreflect.MessageKind {
			child, size := protowire.ConsumeBytes(raw)
			if size < 0 {
				return fmt.Errorf("invalid OTLP protobuf message")
			}
			if err := preflightProtobuf(child, field.Message(), budget, depth+1); err != nil {
				return err
			}
		}
		raw = raw[fieldSize:]
	}
	return nil
}

func preflightJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	depth, containers, tokens := 0, 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scan OTLP JSON: %w", err)
		}
		tokens++
		if tokens > 600000 {
			return fmt.Errorf("OTLP JSON allocation limit exceeded")
		}
		if delimiter, ok := token.(json.Delim); ok {
			if delimiter == '{' || delimiter == '[' {
				depth++
				containers++
			} else {
				depth--
			}
			if depth > 64 || containers > 100000 {
				return fmt.Errorf("OTLP JSON nesting or allocation limit exceeded")
			}
		}
	}
}

func reserveTelemetryBytes(total *int, parts ...string) error {
	for _, part := range parts {
		*total += len(part)
	}
	*total += 1024
	if *total > maxExpandedTelemetryBytes {
		return fmt.Errorf("expanded OTLP storage limit exceeded")
	}
	return nil
}
