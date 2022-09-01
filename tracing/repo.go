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
	"errors"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Repo is a ReadWriteRepo that adds tracing.
type Repo struct {
	eh.ReadWriteRepo
}

// NewRepo creates a new Repo.
func NewRepo(repo eh.ReadWriteRepo) *Repo {
	return &Repo{
		ReadWriteRepo: repo,
	}
}

// InnerRepo implements the InnerRepo method of the eventhorizon.ReadRepo interface.
func (r *Repo) InnerRepo(ctx context.Context) eh.ReadRepo {
	return r.ReadWriteRepo
}

// IntoRepo tries to convert a eh.ReadRepo into a Repo by recursively looking at
// inner repos. Returns nil if none was found.
func IntoRepo(ctx context.Context, repo eh.ReadRepo) *Repo {
	if repo == nil {
		return nil
	}

	if r, ok := repo.(*Repo); ok {
		return r
	}

	return IntoRepo(ctx, repo.InnerRepo(ctx))
}

// Find implements the Find method of the eventhorizon.ReadModel interface.
func (r *Repo) Find(ctx context.Context, id uuid.UUID) (eh.Entity, error) {
	_, span := otel.Tracer("").Start(ctx, "Repo.Find")
	defer span.End()

	entity, err := r.ReadWriteRepo.Find(ctx, id)

	span.SetAttributes(
		attribute.String("eh.aggregate_id", id.String()),
	)

	if err != nil && !errors.Is(err, eh.ErrEntityNotFound) {
		span.RecordError(err)
	}

	return entity, err
}

// FindAll implements the FindAll method of the eventhorizon.ReadRepo interface.
func (r *Repo) FindAll(ctx context.Context) ([]eh.Entity, error) {
	_, span := otel.Tracer("").Start(ctx, "Repo.FindAll")
	defer span.End()

	entities, err := r.ReadWriteRepo.FindAll(ctx)
	if err != nil {
		span.RecordError(err)
	}

	return entities, err
}

// Save implements the Save method of the eventhorizon.WriteRepo interface.
func (r *Repo) Save(ctx context.Context, entity eh.Entity) error {
	_, span := otel.Tracer("").Start(ctx, "Repo.Save")
	defer span.End()

	err := r.ReadWriteRepo.Save(ctx, entity)
	if err != nil {
		span.RecordError(err)
	}

	span.SetAttributes(
		attribute.String("eh.aggregate_id", entity.EntityID().String()),
	)

	return err
}

// Remove implements the Remove method of the eventhorizon.WriteRepo interface.
func (r *Repo) Remove(ctx context.Context, id uuid.UUID) error {
	_, span := otel.Tracer("").Start(ctx, "Repo.Remove")
	defer span.End()

	err := r.ReadWriteRepo.Remove(ctx, id)
	if err != nil {
		span.RecordError(err)
	}

	span.SetAttributes(
		attribute.String("eh.aggregate_id", id.String()),
	)

	return err
}
