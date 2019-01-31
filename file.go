package rpool

import (
	"sync"

	"github.com/moisespsena-go/io-common"
)

// PoolReader is a wrapper around iocommon.ReadSeekCloser to modify the the behavior of
// iocommon.ReadSeekCloser's Close() method.
type PoolReader struct {
	iocommon.ReadSeekCloser
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

// Close() puts the given connects back to the pool instead of closing it.
func (p *PoolReader) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.ReadSeekCloser != nil {
			return p.ReadSeekCloser.Close()
		}
		return nil
	}
	return p.c.put(p.ReadSeekCloser)
}

// MarkUnusable() marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (p *PoolReader) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// newConn wraps a standard iocommon.ReadSeekCloser to a poolConn iocommon.ReadSeekCloser.
func (c *channelPool) wrap(reader iocommon.ReadSeekCloser) iocommon.ReadSeekCloser {
	p := &PoolReader{c: c}
	p.ReadSeekCloser = reader
	return p
}
