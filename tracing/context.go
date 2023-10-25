// Copyright (c) 2020 - The Event Horizon authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"
	"encoding/json"
	"log"

	eh "github.com/looplab/eventhorizon"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// The string keys to marshal the context.
const (
	tracingSpanKeyStr = "eh_tracing_span"
)

// RegisterContext registers the tracing span to be marshaled/unmarshaled on the
// context. This enables propagation of the tracing spans for backends that
// supports it (like Jaeger).
//
// For usage with Elastic APM which doesn't support submitting of child spans
// for the same parent span multiple times outside of a single transaction don't
// register the context. This will provide a new context upon handling in the
// event bus or outbox, which currently is the best Elastic APM can support.
//
// See: https://github.com/elastic/apm/issues/122
//
// OTEL version from @jcampii
func RegisterContext() {
	eh.RegisterContextMarshaler(func(ctx context.Context, vals map[string]interface{}) {
		if span := trace.SpanFromContext(ctx); span != nil {
			propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			carrier := propagation.MapCarrier{}
			propgator.Inject(ctx, &carrier)

			js, err := json.Marshal(carrier)
			if err != nil {
				log.Printf("eventhorizon: could not marshal tracing span: %s", err)

				return
			}

			vals[tracingSpanKeyStr] = string(js)
		}
	})
	eh.RegisterContextUnmarshaler(func(ctx context.Context, vals map[string]interface{}) context.Context {
		if js, ok := vals[tracingSpanKeyStr].(string); ok {
			propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			carrier := propagation.MapCarrier{}
			if err := json.Unmarshal([]byte(js), &carrier); err != nil {
				log.Printf("eventhorizon: could not unmarshal tracing span: %s", err)

				return ctx
			}

			parentCtx := propgator.Extract(ctx, carrier)
			spanCtx := trace.SpanContextFromContext(parentCtx)
			ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
			otel.Tracer("eventbus").Start(ctx, "eventbus", trace.WithSpanKind(trace.SpanKindServer))
		}

		return ctx
	})
}
