package broker

import (
	"context"
	"sync"
)

type Value = string

type Data struct {
	Name  string
	Value Value
}

type Broker interface {
	Write(data Data)
	Read(ctx context.Context, name string) (Value, error)
}

type queue interface {
	Write(value Value)
	Read(ctx context.Context) (Value, error)
}

type BrokerImpl struct {
	mtx    sync.Mutex
	queues map[string]queue
}

type queueImpl struct {
	values      []Value
	valuesCond  *sync.Cond
	readers     []*readChannel
	readersCond *sync.Cond
}

type readChannel struct {
	ch       chan Value
	mtx      sync.Mutex
	isClosed bool
	hasValue bool
}

func newQueue() queue {
	q := &queueImpl{
		valuesCond:  sync.NewCond(&sync.Mutex{}),
		readersCond: sync.NewCond(&sync.Mutex{}),
	}

	go q.read()
	return q
}

func NewBroker() Broker {
	return &BrokerImpl{
		queues: make(map[string]queue),
	}
}

func (qi *queueImpl) Write(value Value) {
	qi.valuesCond.L.Lock()

	qi.values = append(qi.values, value)
	qi.valuesCond.Signal()

	qi.valuesCond.L.Unlock()
}

func (qi *queueImpl) Read(ctx context.Context) (Value, error) {
	readChan := &readChannel{
		ch: make(chan Value, 1),
	}
	qi.readersCond.L.Lock()
	qi.readers = append(qi.readers, readChan)
	qi.readersCond.L.Unlock()
	qi.readersCond.Signal()

	select {
	case <-ctx.Done():
		readChan.mtx.Lock()
		readChan.isClosed = true
		if readChan.hasValue {
			return <-readChan.ch, nil
		}
		close(readChan.ch)
		readChan.mtx.Unlock()
		return "", ctx.Err()
	case v := <-readChan.ch:
		return v, nil
	}
}

func (qi *queueImpl) getValue() Value {
	qi.valuesCond.L.Lock()
	defer qi.valuesCond.L.Unlock()

	for len(qi.values) == 0 {
		qi.valuesCond.Wait()
	}
	value := qi.values[0]

	// не было задачи поработать над производительностью, потому решение вот такое:
	// p.s.: можно заменить на циклический буфер с изменяемой длиной
	copy(qi.values[0:], qi.values[1:])
	qi.values = qi.values[:len(qi.values)-1]

	return value
}

func (qi *queueImpl) getNextReadChan() *readChannel {
	qi.readersCond.L.Lock()
	defer qi.readersCond.L.Unlock()

	for len(qi.readers) == 0 {
		qi.readersCond.Wait()
	}
	reader := qi.readers[0]

	// не было задачи поработать над производительностью, потому решение вот такое:
	// p.s.: можно заменить на циклический буфер с изменяемой длиной
	copy(qi.readers[0:], qi.readers[1:])
	qi.readers = qi.readers[:len(qi.readers)-1]

	return reader
}

func (qi *queueImpl) readValue(value Value) {
	for {
		readChan := qi.getNextReadChan()
		readChan.mtx.Lock()
		if readChan.isClosed {
			continue
		}

		readChan.ch <- value
		readChan.hasValue = true
		readChan.mtx.Unlock()
		return
	}
}

func (qi *queueImpl) read() {
	for {
		value := qi.getValue()
		qi.readValue(value)
	}
}

func (bi *BrokerImpl) getQueue(name string) queue {
	bi.mtx.Lock()
	defer bi.mtx.Unlock()
	q, ok := bi.queues[name]
	if !ok {
		q = newQueue()
		bi.queues[name] = q
	}

	return q
}
func (bi *BrokerImpl) Write(data Data) {
	q := bi.getQueue(data.Name)
	q.Write(data.Value)
}

func (bi *BrokerImpl) Read(ctx context.Context, name string) (Value, error) {
	q := bi.getQueue(name)
	return q.Read(ctx)
}
