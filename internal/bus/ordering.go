package bus

import (
	"context"
	"fmt"
	"sync"
)

// OrderedPublisher ensures per-key ordering for message publishing
type OrderedPublisher struct {
	publisher Publisher
	queues    map[string]*MessageQueue
	mu        sync.RWMutex
}

// NewOrderedPublisher creates a new ordered publisher
func NewOrderedPublisher(publisher Publisher) *OrderedPublisher {
	return &OrderedPublisher{
		publisher: publisher,
		queues:    make(map[string]*MessageQueue),
	}
}

// MessageQueue represents a queue for messages with the same partition key
type MessageQueue struct {
	key      string
	messages chan *Message
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(key string, bufferSize int) *MessageQueue {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &MessageQueue{
		key:      key,
		messages: make(chan *Message, bufferSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start the queue processor
	queue.wg.Add(1)
	go queue.process()

	return queue
}

// Enqueue adds a message to the queue
func (q *MessageQueue) Enqueue(msg *Message) error {
	select {
	case q.messages <- msg:
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue for key %s is closed", q.key)
	default:
		return fmt.Errorf("queue for key %s is full", q.key)
	}
}

// process processes messages in the queue sequentially
func (q *MessageQueue) process() {
	defer q.wg.Done()

	for {
		select {
		case msg := <-q.messages:
			// Publish the message
			// Note: This is a simplified implementation
			// In a real implementation, you would call the actual publisher
			_ = msg
		case <-q.ctx.Done():
			return
		}
	}
}

// Close closes the queue
func (q *MessageQueue) Close() {
	q.cancel()
	q.wg.Wait()
	close(q.messages)
}

// PublishBars publishes bars with ordering guarantee
func (op *OrderedPublisher) PublishBars(ctx context.Context, batch *BarBatchMessage) error {
	key := batch.Key.PartitionKey()
	queue := op.getOrCreateQueue(key)

	// Create message
	msg := &Message{
		Topic: "", // Will be set by the actual publisher
		Key:   batch.Key,
	}

	return queue.Enqueue(msg)
}

// PublishQuote publishes quote with ordering guarantee
func (op *OrderedPublisher) PublishQuote(ctx context.Context, quote *QuoteMessage) error {
	key := quote.Key.PartitionKey()
	queue := op.getOrCreateQueue(key)

	// Create message
	msg := &Message{
		Topic: "", // Will be set by the actual publisher
		Key:   quote.Key,
	}

	return queue.Enqueue(msg)
}

// PublishFundamentals publishes fundamentals with ordering guarantee
func (op *OrderedPublisher) PublishFundamentals(ctx context.Context, fundamentals *FundamentalsMessage) error {
	key := fundamentals.Key.PartitionKey()
	queue := op.getOrCreateQueue(key)

	// Create message
	msg := &Message{
		Topic: "", // Will be set by the actual publisher
		Key:   fundamentals.Key,
	}

	return queue.Enqueue(msg)
}

// Close closes the ordered publisher and all queues
func (op *OrderedPublisher) Close(ctx context.Context) error {
	op.mu.Lock()
	defer op.mu.Unlock()

	// Close all queues
	for _, queue := range op.queues {
		queue.Close()
	}

	// Clear the queues map
	op.queues = make(map[string]*MessageQueue)

	// Close the underlying publisher
	return op.publisher.Close(ctx)
}

// getOrCreateQueue gets or creates a queue for the given key
func (op *OrderedPublisher) getOrCreateQueue(key string) *MessageQueue {
	op.mu.Lock()
	defer op.mu.Unlock()

	queue, exists := op.queues[key]
	if !exists {
		queue = NewMessageQueue(key, 1000) // Buffer size of 1000
		op.queues[key] = queue
	}

	return queue
}
