package consumer

import (
	"context"

	"github.com/hungryhub-team/clickhouse-manager/entity"
	"github.com/hungryhub-team/clickhouse-manager/internal/helper"
)

type ExampleQueue struct {
	ctx context.Context
}

type ExampleConsumer interface {
	Process(payload map[string]interface{}) error
}

func NewExampleConsumer(
	ctx context.Context,
) ExampleConsumer {
	return &ExampleQueue{ctx}
}

func (l *ExampleQueue) Process(payload map[string]interface{}) error {
	var params entity.Log
	params.LoadFromMap(payload)

	helper.Dump(params)

	return nil
}
