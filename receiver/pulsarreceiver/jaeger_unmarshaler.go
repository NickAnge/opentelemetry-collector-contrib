// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pulsarreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/pulsarreceiver"

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	jaegerproto "github.com/jaegertracing/jaeger-idl/model/v1"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger"
)

// copy from kafka receiver
type jaegerProtoSpanUnmarshaler struct{}

var _ TracesUnmarshaler = (*jaegerProtoSpanUnmarshaler)(nil)

func (jaegerProtoSpanUnmarshaler) Unmarshal(bytes []byte) (ptrace.Traces, error) {
	span := &jaegerproto.Span{}
	err := span.Unmarshal(bytes)
	if err != nil {
		return ptrace.NewTraces(), err
	}
	return jaegerSpanToTraces(span)
}

func (jaegerProtoSpanUnmarshaler) Encoding() string {
	return "jaeger_proto"
}

type jaegerJSONSpanUnmarshaler struct{}

var _ TracesUnmarshaler = (*jaegerJSONSpanUnmarshaler)(nil)

func (jaegerJSONSpanUnmarshaler) Unmarshal(data []byte) (ptrace.Traces, error) {
	span := &jaegerproto.Span{}
	err := jsonpb.Unmarshal(bytes.NewReader(data), span)
	if err != nil {
		return ptrace.NewTraces(), err
	}
	return jaegerSpanToTraces(span)
}

func (jaegerJSONSpanUnmarshaler) Encoding() string {
	return "jaeger_json"
}

func jaegerSpanToTraces(span *jaegerproto.Span) (ptrace.Traces, error) {
	batch := jaegerproto.Batch{
		Spans:   []*jaegerproto.Span{span},
		Process: span.Process,
	}
	return jaeger.ProtoToTraces([]*jaegerproto.Batch{&batch})
}
