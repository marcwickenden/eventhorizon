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
	"fmt"

	eh "github.com/looplab/eventhorizon"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// NewCommandHandlerMiddleware returns a new command handler middleware that adds tracing spans.
func NewCommandHandlerMiddleware() eh.CommandHandlerMiddleware {
	return eh.CommandHandlerMiddleware(func(h eh.CommandHandler) eh.CommandHandler {
		return eh.CommandHandlerFunc(func(ctx context.Context, cmd eh.Command) error {
			opName := fmt.Sprintf("Command(%s)", cmd.CommandType())
			_, span := otel.Tracer("").Start(ctx, opName)
			defer span.End()

			err := h.HandleCommand(ctx, cmd)

			span.SetAttributes(
				attribute.String("eh.command_type", cmd.CommandType().String()),
				attribute.String("eh.aggregate_type", cmd.AggregateType().String()),
				attribute.String("eh.aggregate_id", cmd.AggregateID().String()),
			)
			if err != nil {
				span.RecordError(err)
			}

			return err
		})
	})
}
