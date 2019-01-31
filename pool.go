// Package pool implements a pool of iocommon.ReadSeekCloser interfaces to manage and reuse them.
package rpool

import (
	"errors"

	"github.com/moisespsena-go/io-common"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Pool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type Pool interface {
	// Get returns a new reader from the pool. Closing the readers puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error.
	Get() (iocommon.ReadSeekCloser, error)

	// Close closes the pool and all its readers. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of readers of the pool.
	Len() int
}
