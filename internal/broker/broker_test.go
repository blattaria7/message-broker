package broker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Мне просто было удобно разрабатывать эту часть по TDD

func TestBrokerImpl_Read(t *testing.T) {
	data := Data{
		Name:  "name",
		Value: "value",
	}

	broker := NewBroker()
	broker.Write(data)
	value, err := broker.Read(context.Background(), data.Name)

	assert.NoError(t, err)
	assert.Equal(t, data.Value, value)
}

func TestQueueImpl_Read(t *testing.T) {
	q := newQueue()

	q.Write("1")
	value, err := q.Read(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "1", value)
}

func TestQueueImpl_ReadEmpty(t *testing.T) {
	q := newQueue()

	group := sync.WaitGroup{}
	group.Add(1)
	go func() {
		defer group.Done()
		value, err := q.Read(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, "1", value)
	}()

	time.Sleep(time.Second)

	q.Write("1")

	group.Wait()
}

func TestQueueImpl_ReadSome(t *testing.T) {
	q := newQueue()

	q.Write("1")
	q.Write("2")
	q.Write("3")
	q.Write("4")

	value, err := q.Read(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "1", value)

	value, err = q.Read(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "2", value)

	value, err = q.Read(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "3", value)

	value, err = q.Read(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "4", value)
}

func TestBrokerImpl_ReadTimeout(t *testing.T) {
	broker := NewBroker()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	group := sync.WaitGroup{}
	group.Add(1)
	go func() {
		defer group.Done()
		v, err := broker.Read(ctx, "name")
		assert.Empty(t, v)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}()

	group.Wait()
}

func TestBrokerImpl_ReadTimeoutSome(t *testing.T) {
	broker := NewBroker()

	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)

	group := sync.WaitGroup{}
	group.Add(1)
	go func() {
		defer group.Done()
		v, err := broker.Read(context.Background(), "name")
		assert.Equal(t, "1", v)
		assert.NoError(t, err)

		v, err = broker.Read(ctx, "name")
		assert.Empty(t, v)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		v, err = broker.Read(context.Background(), "name")
		assert.Equal(t, "2", v)
		assert.NoError(t, err)

		v, err = broker.Read(context.Background(), "name")
		assert.Equal(t, "3", v)
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second)

	broker.Write(Data{"name", "1"})
	broker.Write(Data{"name", "2"})
	broker.Write(Data{"name", "3"})

	group.Wait()
}
