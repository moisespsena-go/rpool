package rpool

import (
	"errors"
	"fmt"
	"sync"

	"github.com/moisespsena-go/io-common"
)

// channelPool implements the Pool interface based on buffered channels.
type channelPool struct {
	// storage for our iocommon.ReadSeekCloser connections
	mu      sync.RWMutex
	readers chan iocommon.ReadSeekCloser

	// iocommon.ReadSeekCloser generator
	factory Factory
}

// Factory is a function to create new connections.
type Factory func() (iocommon.ReadSeekCloser, error)

// NewChannelPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewChannelPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		readers: make(chan iocommon.ReadSeekCloser, maxCap),
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.readers <- conn
	}

	return c, nil
}

func (c *channelPool) getConnsAndFactory() (chan iocommon.ReadSeekCloser, Factory) {
	c.mu.RLock()
	readers := c.readers
	factory := c.factory
	c.mu.RUnlock()
	return readers, factory
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (iocommon.ReadSeekCloser, error) {
	readers, factory := c.getConnsAndFactory()
	if readers == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom iocommon.ReadSeekCloser implementation (wrap
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-readers:
		if conn == nil {
			return nil, ErrClosed
		}

		return c.wrap(conn), nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		return c.wrap(conn), nil
	}
}

// put puts the reader back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) put(conn iocommon.ReadSeekCloser) error {
	if conn == nil {
		return errors.New("reader is nil. rejecting")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.readers == nil {
		// pool is closed, close passed reader
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.readers <- conn:
		return nil
	default:
		// pool is full, close passed reader
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	readers := c.readers
	c.readers = nil
	c.factory = nil
	c.mu.Unlock()

	if readers == nil {
		return
	}

	close(readers)
	for conn := range readers {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	readers, _ := c.getConnsAndFactory()
	return len(readers)
}
