package rpool

import (
	"testing"

	"github.com/moisespsena-go/io-common"
)

func TestConn_Impl(t *testing.T) {
	var _ iocommon.ReadSeekCloser = new(PoolReader)
}
