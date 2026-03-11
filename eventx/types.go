package eventx

import (
	"context"
	"reflect"

	"github.com/DaiYuANg/arcgo/collectionx"
)

// Event is the common event contract for strongly typed publish/subscribe.
type Event interface {
	Name() string
}

type subscription struct {
	id      uint64
	handler HandlerFunc
}

type publishTask struct {
	ctx      context.Context
	event    Event
	handlers []HandlerFunc
}

// subscriptionTable is a concurrent table for storing subscriptions by (event type, subscription id).
type subscriptionTable = collectionx.ConcurrentTable[reflect.Type, uint64, *subscription]
