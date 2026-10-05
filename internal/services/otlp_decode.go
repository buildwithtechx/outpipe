package services

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const MaxOTLPBodyBytes = 4 * 1024 * 1024
const maxOTLPRecords = 10000

func decodeOTLP(raw []byte, contentType string, message proto.Message) error {
	if len(raw) > MaxOTLPBodyBytes {
		return fmt.Errorf("OTLP payload must contain between 1 and %d bytes", MaxOTLPBodyBytes)
	}
	typeName, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("parse OTLP content type: %w", err)
	}
	if typeName == "application/x-protobuf" {
		if err := preflightProtobuf(raw, message.ProtoReflect().Descriptor(), &otlpParseBudget{}, 0); err != nil {
			return err
		}
		if err := (proto.UnmarshalOptions{RecursionLimit: 64}).Unmarshal(raw, message); err != nil {
			return fmt.Errorf("decode OTLP protobuf: %w", err)
		}
		return nil
	}
	if typeName != "application/json" {
		return fmt.Errorf("unsupported OTLP content type")
	}
	if err := preflightJSON(raw); err != nil {
		return err
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode OTLP JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("OTLP JSON must contain one message")
	}
	if err := normalizeOTLPIDs(value, 0); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode normalized OTLP JSON: %w", err)
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true, RecursionLimit: 64}).Unmarshal(data, message); err != nil {
		return fmt.Errorf("decode OTLP JSON message: %w", err)
	}
	return nil
}

func normalizeOTLPIDs(value any, depth int) error {
	if depth > 64 {
		return fmt.Errorf("OTLP JSON nesting exceeds limit")
	}
	switch value := value.(type) {
	case map[string]any:
		for key, item := range value {
			if key == "traceId" || key == "spanId" || key == "parentSpanId" {
				text, ok := item.(string)
				if !ok {
					return fmt.Errorf("OTLP identifier must be a hexadecimal string")
				}
				decoded, err := hex.DecodeString(text)
				if err != nil {
					return fmt.Errorf("decode OTLP identifier: %w", err)
				}
				if len(decoded) != 0 && ((key == "traceId" && len(decoded) != 16) || (key != "traceId" && len(decoded) != 8)) {
					return fmt.Errorf("invalid OTLP identifier length")
				}
				value[key] = base64.StdEncoding.EncodeToString(decoded)
			} else if err := normalizeOTLPIDs(item, depth+1); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range value {
			if err := normalizeOTLPIDs(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}
