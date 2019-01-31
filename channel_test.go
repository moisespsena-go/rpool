package rpool

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/moisespsena-go/io-common"
)

var (
	InitialCap = 5
	MaximumCap = 30
	testData   = []byte("test data")
	factory    = func() (iocommon.ReadSeekCloser, error) {
		return iocommon.NewBytesReadCloser(testData), nil
	}
)

func TestNew(t *testing.T) {
	_, err := newChannelPool()
	if err != nil {
		t.Errorf("New error: %s", err)
	}
}
func TestPool_Get_Impl(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	reader, err := p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

	_, ok := reader.(*PoolReader)
	if !ok {
		t.Errorf("Reader is not of type poolReader")
	}
}

func TestPool_Get(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	_, err := p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

	// after one get, current capacity should be lowered by one.
	if p.Len() != (InitialCap - 1) {
		t.Errorf("Get error. Expecting %d, got %d",
			(InitialCap - 1), p.Len())
	}

	// get them all
	var wg sync.WaitGroup
	for i := 0; i < (InitialCap - 1); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			if err != nil {
				t.Errorf("Get error: %s", err)
			}
		}()
	}
	wg.Wait()

	if p.Len() != 0 {
		t.Errorf("Get error. Expecting %d, got %d",
			(InitialCap - 1), p.Len())
	}

	_, err = p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
}

func TestPool_Put(t *testing.T) {
	p, err := NewChannelPool(0, 30, factory)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// get/create from the pool
	readers := make([]iocommon.ReadSeekCloser, MaximumCap)
	for i := 0; i < MaximumCap; i++ {
		reader, _ := p.Get()
		readers[i] = reader
	}

	// now put them all back
	for _, reader := range readers {
		reader.Close()
	}

	if p.Len() != MaximumCap {
		t.Errorf("Put error len. Expecting %d, got %d",
			1, p.Len())
	}

	reader, _ := p.Get()
	p.Close() // close pool

	reader.Close() // try to put into a full pool
	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put readerections.")
	}
}

func TestPool_PutUnusableReader(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	// ensure pool is not empty
	reader, _ := p.Get()
	reader.Close()

	poolSize := p.Len()
	reader, _ = p.Get()
	reader.Close()
	if p.Len() != poolSize {
		t.Errorf("Pool size is expected to be equal to initial size")
	}

	reader, _ = p.Get()
	if pc, ok := reader.(*PoolReader); !ok {
		t.Errorf("impossible")
	} else {
		pc.MarkUnusable()
	}
	reader.Close()
	if p.Len() != poolSize-1 {
		t.Error("Pool size is expected to be initial_size - 1", p.Len(), poolSize-1)
	}
}

func TestPool_UsedCapacity(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	if p.Len() != InitialCap {
		t.Errorf("InitialCap error. Expecting %d, got %d",
			InitialCap, p.Len())
	}
}

func TestPool_Close(t *testing.T) {
	p, _ := newChannelPool()

	// now close it and test all cases we are expecting.
	p.Close()

	c := p.(*channelPool)

	if c.readers != nil {
		t.Errorf("Close error, readers channel should be nil")
	}

	if c.factory != nil {
		t.Errorf("Close error, factory should be nil")
	}

	_, err := p.Get()
	if err == nil {
		t.Errorf("Close error, get reader should return an error")
	}

	if p.Len() != 0 {
		t.Errorf("Close error used capacity. Expecting 0, got %d", p.Len())
	}
}

func TestPoolConcurrent(t *testing.T) {
	p, _ := newChannelPool()
	pipe := make(chan iocommon.ReadSeekCloser, 0)

	go func() {
		p.Close()
	}()

	for i := 0; i < MaximumCap; i++ {
		go func() {
			reader, _ := p.Get()

			pipe <- reader
		}()

		go func() {
			reader := <-pipe
			if reader == nil {
				return
			}
			reader.Close()
		}()
	}
}

func TestPoolWriteRead(t *testing.T) {
	p, _ := NewChannelPool(0, 30, factory)

	reader, _ := p.Get()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(data, testData) != 0 {
		t.Errorf("Unexpected data")
	}
}

func TestPoolConcurrent2(t *testing.T) {
	p, _ := NewChannelPool(0, 30, factory)

	var wg sync.WaitGroup

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				reader, _ := p.Get()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				reader.Close()
				wg.Done()
			}(i)
		}
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			reader, _ := p.Get()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			reader.Close()
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestPoolConcurrent3(t *testing.T) {
	p, _ := NewChannelPool(0, 1, factory)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		p.Close()
		wg.Done()
	}()

	if reader, err := p.Get(); err == nil {
		reader.Close()
	}

	wg.Wait()
}

func newChannelPool() (Pool, error) {
	return NewChannelPool(InitialCap, MaximumCap, factory)
}
