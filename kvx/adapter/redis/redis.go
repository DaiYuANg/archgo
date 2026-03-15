// Package redis provides a Redis adapter for kvx.
package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DaiYuANg/archgo/kvx"
	"github.com/redis/go-redis/v9"
)

// Adapter implements kvx.Client using go-redis.
type Adapter struct {
	client *redis.Client
}

// New creates a new Redis adapter.
func New(opts kvx.ClientOptions) (*Adapter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            opts.Addrs[0],
		Password:        opts.Password,
		DB:              opts.DB,
		TLSConfig:       nil, // TODO: support TLS
		PoolSize:        opts.PoolSize,
		MinIdleConns:    opts.MinIdleConns,
		ConnMaxLifetime: opts.ConnMaxLifetime,
		ConnMaxIdleTime: opts.ConnMaxIdleTime,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Adapter{client: rdb}, nil
}

// NewFromClient creates an adapter from an existing redis.Client.
func NewFromClient(client *redis.Client) *Adapter {
	return &Adapter{client: client}
}

// Close closes the client connection.
func (a *Adapter) Close() error {
	return a.client.Close()
}

// ============== KV Interface ==============

// Get retrieves the value for the given key.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := a.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, kvx.ErrNil
		}
		return nil, err
	}
	return []byte(val), nil
}

// MGet retrieves multiple values for the given keys.
func (a *Adapter) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	vals, err := a.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, v := range vals {
		if v != nil {
			if str, ok := v.(string); ok {
				result[keys[i]] = []byte(str)
			}
		}
	}
	return result, nil
}

// Set sets the value for the given key.
func (a *Adapter) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return a.client.Set(ctx, key, value, expiration).Err()
}

// MSet sets multiple key-value pairs.
func (a *Adapter) MSet(ctx context.Context, values map[string][]byte, expiration time.Duration) error {
	// Use MSet for atomic operation
	ifaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		ifaceValues[k] = v
	}

	if err := a.client.MSet(ctx, ifaceValues).Err(); err != nil {
		return err
	}

	// Set expiration if needed
	if expiration > 0 {
		for key := range values {
			if err := a.client.Expire(ctx, key, expiration).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete deletes the given key.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	return a.client.Del(ctx, key).Err()
}

// DeleteMulti deletes multiple keys.
func (a *Adapter) DeleteMulti(ctx context.Context, keys []string) error {
	return a.client.Del(ctx, keys...).Err()
}

// Exists checks if the key exists.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	n, err := a.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ExistsMulti checks if multiple keys exist.
func (a *Adapter) ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error) {
	results := make(map[string]bool, len(keys))
	for _, key := range keys {
		exists, err := a.Exists(ctx, key)
		if err != nil {
			return nil, err
		}
		results[key] = exists
	}
	return results, nil
}

// Expire sets the expiration for the given key.
func (a *Adapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.client.Expire(ctx, key, expiration).Err()
}

// TTL gets the TTL for the given key.
func (a *Adapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.client.TTL(ctx, key).Result()
}

// Scan iterates over keys matching the pattern.
func (a *Adapter) Scan(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	return a.client.Scan(ctx, cursor, pattern, count).Result()
}

// Keys returns all keys matching the pattern.
func (a *Adapter) Keys(ctx context.Context, pattern string) ([]string, error) {
	return a.client.Keys(ctx, pattern).Result()
}

// ============== Hash Interface ==============

// HGet gets a field from a hash.
func (a *Adapter) HGet(ctx context.Context, key string, field string) ([]byte, error) {
	val, err := a.client.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, kvx.ErrNil
		}
		return nil, err
	}
	return []byte(val), nil
}

// HMGet gets multiple fields from a hash.
func (a *Adapter) HMGet(ctx context.Context, key string, fields []string) (map[string][]byte, error) {
	vals, err := a.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, v := range vals {
		if v != nil {
			if str, ok := v.(string); ok {
				result[fields[i]] = []byte(str)
			}
		}
	}
	return result, nil
}

// HSet sets fields in a hash.
func (a *Adapter) HSet(ctx context.Context, key string, values map[string][]byte) error {
	// Convert map[string][]byte to map[string]interface{}
	ifaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		ifaceValues[k] = v
	}
	return a.client.HSet(ctx, key, ifaceValues).Err()
}

// HMSet sets multiple fields in a hash.
func (a *Adapter) HMSet(ctx context.Context, key string, values map[string][]byte) error {
	ifaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		ifaceValues[k] = v
	}
	return a.client.HMSet(ctx, key, ifaceValues).Err()
}

// HGetAll gets all fields and values from a hash.
func (a *Adapter) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	val, err := a.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte, len(val))
	for k, v := range val {
		result[k] = []byte(v)
	}
	return result, nil
}

// HDel deletes fields from a hash.
func (a *Adapter) HDel(ctx context.Context, key string, fields ...string) error {
	return a.client.HDel(ctx, key, fields...).Err()
}

// HExists checks if a field exists in a hash.
func (a *Adapter) HExists(ctx context.Context, key string, field string) (bool, error) {
	return a.client.HExists(ctx, key, field).Result()
}

// HKeys gets all field names in a hash.
func (a *Adapter) HKeys(ctx context.Context, key string) ([]string, error) {
	return a.client.HKeys(ctx, key).Result()
}

// HVals gets all values in a hash.
func (a *Adapter) HVals(ctx context.Context, key string) ([][]byte, error) {
	vals, err := a.client.HVals(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	result := make([][]byte, len(vals))
	for i, v := range vals {
		result[i] = []byte(v)
	}
	return result, nil
}

// HLen gets the number of fields in a hash.
func (a *Adapter) HLen(ctx context.Context, key string) (int64, error) {
	return a.client.HLen(ctx, key).Result()
}

// HIncrBy increments a field by the given value.
func (a *Adapter) HIncrBy(ctx context.Context, key string, field string, increment int64) (int64, error) {
	return a.client.HIncrBy(ctx, key, field, increment).Result()
}

// ============== PubSub Interface ==============

// Publish publishes a message to a channel.
func (a *Adapter) Publish(ctx context.Context, channel string, message []byte) error {
	return a.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel.
func (a *Adapter) Subscribe(ctx context.Context, channel string) (kvx.Subscription, error) {
	pubsub := a.client.Subscribe(ctx, channel)
	// Verify subscription
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return &redisSubscription{pubsub: pubsub}, nil
}

// PSubscribe subscribes to channels matching a pattern.
func (a *Adapter) PSubscribe(ctx context.Context, pattern string) (kvx.Subscription, error) {
	pubsub := a.client.PSubscribe(ctx, pattern)
	// Verify subscription
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return &redisSubscription{pubsub: pubsub}, nil
}

type redisSubscription struct {
	pubsub *redis.PubSub
	once   sync.Once
	ch     chan []byte
}

func (s *redisSubscription) Channel() <-chan []byte {
	s.once.Do(func() {
		s.ch = make(chan []byte, 100)
		go func() {
			defer close(s.ch)
			ch := s.pubsub.Channel()
			for msg := range ch {
				s.ch <- []byte(msg.Payload)
			}
		}()
	})
	return s.ch
}

func (s *redisSubscription) Close() error {
	return s.pubsub.Close()
}

// ============== Stream Interface ==============

// XAdd adds an entry to a stream.
func (a *Adapter) XAdd(ctx context.Context, key string, id string, values map[string][]byte) (string, error) {
	// Convert map[string][]byte to map[string]interface{}
	ifaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		ifaceValues[k] = v
	}

	args := &redis.XAddArgs{
		Stream: key,
		Values: ifaceValues,
	}
	if id != "*" {
		args.ID = id
	}

	return a.client.XAdd(ctx, args).Result()
}

// XRead reads entries from a stream.
func (a *Adapter) XRead(ctx context.Context, key string, start string, count int64) ([]kvx.StreamEntry, error) {
	streams := []string{key, start}

	result, err := a.client.XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   0,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	entries := make([]kvx.StreamEntry, len(result[0].Messages))
	for i, msg := range result[0].Messages {
		entries[i] = kvx.StreamEntry{
			ID:     msg.ID,
			Values: convertInterfaceMapToBytes(msg.Values),
		}
	}
	return entries, nil
}

// XReadMultiple reads entries from multiple streams.
func (a *Adapter) XReadMultiple(ctx context.Context, streams map[string]string, count int64, block time.Duration) (map[string][]kvx.StreamEntry, error) {
	streamKeys := make([]string, 0, len(streams)*2)
	for key, start := range streams {
		streamKeys = append(streamKeys, key, start)
	}

	result, err := a.client.XRead(ctx, &redis.XReadArgs{
		Streams: streamKeys,
		Count:   count,
		Block:   block,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return make(map[string][]kvx.StreamEntry), nil
		}
		return nil, err
	}

	entries := make(map[string][]kvx.StreamEntry)
	for _, stream := range result {
		streamEntries := make([]kvx.StreamEntry, len(stream.Messages))
		for i, msg := range stream.Messages {
			streamEntries[i] = kvx.StreamEntry{
				ID:     msg.ID,
				Values: convertInterfaceMapToBytes(msg.Values),
			}
		}
		entries[stream.Stream] = streamEntries
	}
	return entries, nil
}

// XRange reads entries in a range.
func (a *Adapter) XRange(ctx context.Context, key string, start, stop string) ([]kvx.StreamEntry, error) {
	result, err := a.client.XRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]kvx.StreamEntry, len(result))
	for i, msg := range result {
		entries[i] = kvx.StreamEntry{
			ID:     msg.ID,
			Values: convertInterfaceMapToBytes(msg.Values),
		}
	}
	return entries, nil
}

// XRevRange reads entries in reverse order.
func (a *Adapter) XRevRange(ctx context.Context, key string, start, stop string) ([]kvx.StreamEntry, error) {
	result, err := a.client.XRevRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]kvx.StreamEntry, len(result))
	for i, msg := range result {
		entries[i] = kvx.StreamEntry{
			ID:     msg.ID,
			Values: convertInterfaceMapToBytes(msg.Values),
		}
	}
	return entries, nil
}

// XLen gets the number of entries in a stream.
func (a *Adapter) XLen(ctx context.Context, key string) (int64, error) {
	return a.client.XLen(ctx, key).Result()
}

// XTrim trims the stream to approximately maxLen entries.
func (a *Adapter) XTrim(ctx context.Context, key string, maxLen int64) error {
	return a.client.XTrimMaxLen(ctx, key, maxLen).Err()
}

// XDel deletes specific entries from a stream.
func (a *Adapter) XDel(ctx context.Context, key string, ids []string) error {
	return a.client.XDel(ctx, key, ids...).Err()
}

// XGroupCreate creates a consumer group.
func (a *Adapter) XGroupCreate(ctx context.Context, key string, group string, startID string) error {
	return a.client.XGroupCreate(ctx, key, group, startID).Err()
}

// XGroupDestroy destroys a consumer group.
func (a *Adapter) XGroupDestroy(ctx context.Context, key string, group string) error {
	return a.client.XGroupDestroy(ctx, key, group).Err()
}

// XGroupCreateConsumer creates a consumer in a group.
func (a *Adapter) XGroupCreateConsumer(ctx context.Context, key string, group string, consumer string) error {
	return a.client.XGroupCreateConsumer(ctx, key, group, consumer).Err()
}

// XGroupDelConsumer deletes a consumer from a group.
func (a *Adapter) XGroupDelConsumer(ctx context.Context, key string, group string, consumer string) error {
	return a.client.XGroupDelConsumer(ctx, key, group, consumer).Err()
}

// XReadGroup reads entries as part of a consumer group.
func (a *Adapter) XReadGroup(ctx context.Context, group string, consumer string, streams map[string]string, count int64, block time.Duration) (map[string][]kvx.StreamEntry, error) {
	streamKeys := make([]string, 0, len(streams)*2)
	for key, start := range streams {
		streamKeys = append(streamKeys, key, start)
	}

	result, err := a.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streamKeys,
		Count:    count,
		Block:    block,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return make(map[string][]kvx.StreamEntry), nil
		}
		return nil, err
	}

	entries := make(map[string][]kvx.StreamEntry)
	for _, stream := range result {
		streamEntries := make([]kvx.StreamEntry, len(stream.Messages))
		for i, msg := range stream.Messages {
			streamEntries[i] = kvx.StreamEntry{
				ID:     msg.ID,
				Values: convertInterfaceMapToBytes(msg.Values),
			}
		}
		entries[stream.Stream] = streamEntries
	}
	return entries, nil
}

// XAck acknowledges processing of stream entries.
func (a *Adapter) XAck(ctx context.Context, key string, group string, ids []string) error {
	return a.client.XAck(ctx, key, group, ids...).Err()
}

// XPending gets pending entries information.
func (a *Adapter) XPending(ctx context.Context, key string, group string) (*kvx.PendingInfo, error) {
	result, err := a.client.XPending(ctx, key, group).Result()
	if err != nil {
		return nil, err
	}

	return &kvx.PendingInfo{
		Count:     result.Count,
		StartID:   result.Lower,
		EndID:     result.Higher,
		Consumers: result.Consumers,
	}, nil
}

// XPendingRange gets pending entries in a range.
func (a *Adapter) XPendingRange(ctx context.Context, key string, group string, start string, stop string, count int64) ([]kvx.PendingEntry, error) {
	result, err := a.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: key,
		Group:  group,
		Start:  start,
		End:    stop,
		Count:  count,
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]kvx.PendingEntry, len(result))
	for i, p := range result {
		entries[i] = kvx.PendingEntry{
			ID:         p.ID,
			Consumer:   p.Consumer,
			IdleTime:   p.Idle,
			Deliveries: p.RetryCount,
		}
	}
	return entries, nil
}

// XClaim claims pending entries for a consumer.
func (a *Adapter) XClaim(ctx context.Context, key string, group string, consumer string, minIdleTime time.Duration, ids []string) ([]kvx.StreamEntry, error) {
	result, err := a.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   key,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdleTime,
		Messages: ids,
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]kvx.StreamEntry, len(result))
	for i, msg := range result {
		entries[i] = kvx.StreamEntry{
			ID:     msg.ID,
			Values: convertInterfaceMapToBytes(msg.Values),
		}
	}
	return entries, nil
}

// XAutoClaim auto-claims pending entries.
func (a *Adapter) XAutoClaim(ctx context.Context, key string, group string, consumer string, minIdleTime time.Duration, start string, count int64) (string, []kvx.StreamEntry, error) {
	messages, next, err := a.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   key,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdleTime,
		Start:    start,
		Count:    count,
	}).Result()
	if err != nil {
		return "", nil, err
	}

	entries := make([]kvx.StreamEntry, len(messages))
	for i, msg := range messages {
		entries[i] = kvx.StreamEntry{
			ID:     msg.ID,
			Values: convertInterfaceMapToBytes(msg.Values),
		}
	}
	return next, entries, nil
}

// XInfoGroups gets info about consumer groups.
func (a *Adapter) XInfoGroups(ctx context.Context, key string) ([]kvx.GroupInfo, error) {
	result, err := a.client.XInfoGroups(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	groups := make([]kvx.GroupInfo, len(result))
	for i, g := range result {
		groups[i] = kvx.GroupInfo{
			Name:            g.Name,
			Consumers:       g.Consumers,
			Pending:         g.Pending,
			LastDeliveredID: g.LastDeliveredID,
		}
	}
	return groups, nil
}

// XInfoConsumers gets info about consumers in a group.
func (a *Adapter) XInfoConsumers(ctx context.Context, key string, group string) ([]kvx.ConsumerInfo, error) {
	result, err := a.client.XInfoConsumers(ctx, key, group).Result()
	if err != nil {
		return nil, err
	}

	consumers := make([]kvx.ConsumerInfo, len(result))
	for i, c := range result {
		consumers[i] = kvx.ConsumerInfo{
			Name:    c.Name,
			Pending: c.Pending,
			Idle:    c.Idle,
		}
	}
	return consumers, nil
}

// XInfoStream gets info about a stream.
func (a *Adapter) XInfoStream(ctx context.Context, key string) (*kvx.StreamInfo, error) {
	result, err := a.client.XInfoStream(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	info := &kvx.StreamInfo{
		Length:          result.Length,
		RadixTreeKeys:   result.RadixTreeKeys,
		RadixTreeNodes:  result.RadixTreeNodes,
		Groups:          result.Groups,
		LastGeneratedID: result.LastGeneratedID,
	}

	if result.FirstEntry.ID != "" {
		info.FirstEntry = &kvx.StreamEntry{
			ID:     result.FirstEntry.ID,
			Values: convertInterfaceMapToBytes(result.FirstEntry.Values),
		}
	}

	if result.LastEntry.ID != "" {
		info.LastEntry = &kvx.StreamEntry{
			ID:     result.LastEntry.ID,
			Values: convertInterfaceMapToBytes(result.LastEntry.Values),
		}
	}

	return info, nil
}

// ============== Script Interface ==============

// Load loads a script into the script cache.
func (a *Adapter) Load(ctx context.Context, script string) (string, error) {
	return a.client.ScriptLoad(ctx, script).Result()
}

// Eval executes a script.
func (a *Adapter) Eval(ctx context.Context, script string, keys []string, args [][]byte) ([]byte, error) {
	ifaceArgs := make([]interface{}, len(args))
	for i, v := range args {
		ifaceArgs[i] = v
	}

	val, err := a.client.Eval(ctx, script, keys, ifaceArgs...).Result()
	if err != nil {
		return nil, err
	}

	return valueToBytes(val)
}

// EvalSHA executes a cached script by SHA.
func (a *Adapter) EvalSHA(ctx context.Context, sha string, keys []string, args [][]byte) ([]byte, error) {
	ifaceArgs := make([]interface{}, len(args))
	for i, v := range args {
		ifaceArgs[i] = v
	}

	val, err := a.client.EvalSha(ctx, sha, keys, ifaceArgs...).Result()
	if err != nil {
		return nil, err
	}

	return valueToBytes(val)
}

// ============== JSON Interface ==============

// JSONSet sets a JSON value at key.
func (a *Adapter) JSONSet(ctx context.Context, key string, path string, value []byte, expiration time.Duration) error {
	// Use JSON.SET command via Do
	err := a.client.Do(ctx, "JSON.SET", key, path, value).Err()
	if err != nil {
		return err
	}

	if expiration > 0 {
		return a.client.Expire(ctx, key, expiration).Err()
	}
	return nil
}

// JSONGet gets a JSON value at key.
func (a *Adapter) JSONGet(ctx context.Context, key string, path string) ([]byte, error) {
	val, err := a.client.Do(ctx, "JSON.GET", key, path).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, kvx.ErrNil
		}
		return nil, err
	}

	return valueToBytes(val)
}

// JSONSetField sets a field in a JSON document.
func (a *Adapter) JSONSetField(ctx context.Context, key string, path string, value []byte) error {
	return a.client.Do(ctx, "JSON.SET", key, path, value).Err()
}

// JSONGetField gets a field from a JSON document.
func (a *Adapter) JSONGetField(ctx context.Context, key string, path string) ([]byte, error) {
	return a.JSONGet(ctx, key, path)
}

// JSONDelete deletes a JSON value or field.
func (a *Adapter) JSONDelete(ctx context.Context, key string, path string) error {
	return a.client.Do(ctx, "JSON.DEL", key, path).Err()
}

// ============== Search Interface ==============

// CreateIndex creates a secondary index.
func (a *Adapter) CreateIndex(ctx context.Context, indexName string, prefix string, schema []kvx.SchemaField) error {
	args := make([]interface{}, 0)
	args = append(args, indexName, "ON", "HASH", "PREFIX", 1, prefix, "SCHEMA")

	for _, f := range schema {
		args = append(args, f.Name, string(f.Type))
		if f.Sortable {
			args = append(args, "SORTABLE")
		}
	}

	return a.client.Do(ctx, args...).Err()
}

// DropIndex drops a secondary index.
func (a *Adapter) DropIndex(ctx context.Context, indexName string) error {
	return a.client.Do(ctx, "FT.DROPINDEX", indexName).Err()
}

// Search performs a search query.
func (a *Adapter) Search(ctx context.Context, indexName string, query string, limit int) ([]string, error) {
	val, err := a.client.Do(ctx, "FT.SEARCH", indexName, query, "LIMIT", 0, limit).Result()
	if err != nil {
		return nil, err
	}

	// Parse FT.SEARCH response
	// Response format: [total, key1, [field1, value1, ...], key2, ...]
	return parseFTSearchResponse(val)
}

// SearchWithSort performs a search query with sorting.
func (a *Adapter) SearchWithSort(ctx context.Context, indexName string, query string, sortBy string, ascending bool, limit int) ([]string, error) {
	args := []interface{}{"FT.SEARCH", indexName, query, "SORTBY", sortBy}
	if !ascending {
		args = append(args, "DESC")
	}
	args = append(args, "LIMIT", 0, limit)

	val, err := a.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, err
	}

	return parseFTSearchResponse(val)
}

// SearchAggregate performs an aggregation query.
func (a *Adapter) SearchAggregate(ctx context.Context, indexName string, query string, limit int) ([]map[string]interface{}, error) {
	val, err := a.client.Do(ctx, "FT.AGGREGATE", indexName, query, "LIMIT", 0, limit).Result()
	if err != nil {
		return nil, err
	}

	// Parse FT.AGGREGATE response
	return parseFTAggregateResponse(val)
}

// ============== Pipeline Interface ==============

// Pipeline creates a new pipeline.
func (a *Adapter) Pipeline() kvx.Pipeline {
	return &redisPipeline{
		pipe: a.client.Pipeline(),
	}
}

type redisPipeline struct {
	pipe redis.Pipeliner
}

// Enqueue adds a command to the pipeline.
func (p *redisPipeline) Enqueue(command string, args ...[]byte) {
	// Convert args to interface{}
	ifaceArgs := make([]interface{}, len(args)+1)
	ifaceArgs[0] = command
	for i, v := range args {
		ifaceArgs[i+1] = v
	}
	p.pipe.Do(context.Background(), ifaceArgs...)
}

// Exec executes all queued commands.
func (p *redisPipeline) Exec(ctx context.Context) ([][]byte, error) {
	cmders, err := p.pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	results := make([][]byte, len(cmders))
	for i, cmd := range cmders {
		val, err := cmd.(*redis.Cmd).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			results[i] = nil
			continue
		}
		results[i], _ = valueToBytes(val)
	}
	return results, nil
}

// Close closes the pipeline.
func (p *redisPipeline) Close() error {
	// Pipeline doesn't need explicit close in go-redis
	return nil
}

// ============== Lock Interface ==============

// Acquire tries to acquire a lock.
func (a *Adapter) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use SET NX for simple distributed lock
	ok, err := a.client.SetNX(ctx, key, "1", ttl).Result()
	return ok, err
}

// Release releases a lock.
func (a *Adapter) Release(ctx context.Context, key string) error {
	return a.client.Del(ctx, key).Err()
}

// Extend extends the lock TTL.
func (a *Adapter) Extend(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use PEXPIRE to extend the lock
	ok, err := a.client.Expire(ctx, key, ttl).Result()
	return ok, err
}

// ============== Helper Functions ==============

func convertInterfaceMapToBytes(m map[string]interface{}) map[string][]byte {
	result := make(map[string][]byte, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case []byte:
			result[k] = val
		case string:
			result[k] = []byte(val)
		default:
			result[k] = []byte(fmt.Sprintf("%v", val))
		}
	}
	return result
}

func valueToBytes(val interface{}) ([]byte, error) {
	switch v := val.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case nil:
		return nil, nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}

func parseFTSearchResponse(val interface{}) ([]string, error) {
	arr, ok := val.([]interface{})
	if !ok {
		return nil, nil
	}

	if len(arr) < 1 {
		return nil, nil
	}

	// Extract keys from the response
	var keys []string
	for i := 1; i < len(arr); i += 2 {
		if key, ok := arr[i].(string); ok {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func parseFTAggregateResponse(val interface{}) ([]map[string]interface{}, error) {
	arr, ok := val.([]interface{})
	if !ok {
		return nil, nil
	}

	if len(arr) < 1 {
		return nil, nil
	}

	// Parse aggregation results
	var results []map[string]interface{}
	for i := 1; i < len(arr); i++ {
		if row, ok := arr[i].([]interface{}); ok {
			result := make(map[string]interface{})
			for j := 0; j < len(row)-1; j += 2 {
				if key, ok := row[j].(string); ok {
					result[key] = row[j+1]
				}
			}
			results = append(results, result)
		}
	}
	return results, nil
}
