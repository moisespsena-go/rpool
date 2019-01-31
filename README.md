# Archived project. No maintenance. 
This project is not maintained anymore and is archived. Feel free to fork and
use make your own changes if needed.

Thanks all for their work on this project. 

# RPool [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/moisespsena-go/rpool) [![Build Status](http://img.shields.io/travis/moisespsena-go/rpool.svg?style=flat-square)](https://travis-ci.org/moisespsena-go/rpool)

Reader Pool is a thread safe reader pool for iocommon.ReadSeekCloser interface. It can be used to
manage and reuse readers.


## Install and Usage

Install the package with:

```bash
go get github.com/moisespsena-go/rpool
```

Please vendor the package with one of the releases: https://github.com/moisespsena-go/rpool/releases.
`master` branch is **development** branch and will contain always the latest changes.


## Example

```go
// create a factory() to be used with channel based pool
factory    := func() (iocommon.ReadSeekCloser, error) { return os.Open("file.txt") }

// create a new channel based pool with an initial capacity of 5 and maximum
// capacity of 30. The factory will create 5 initial readers and put it
// into the pool.
p, err := rpool.NewChannelPool(5, 30, factory)

// now you can get a reader from the pool, if there is no reader
// available it will create a new one via the factory function.
reader, err := p.Get()

// do something with reader and put it back to the pool by closing the reader
// (this doesn't close the underlying reader instead it's putting it back
// to the pool).
reader.Close()

// close the underlying reader instead of returning it to pool
// it is useful when acceptor has already closed reader and reader.Write() returns error
if pr, ok := reader.(*rpool.PoolReader); ok {
  pr.MarkUnusable()
  pr.Close()
}

// close pool any time you want, this closes all the readers inside a pool
p.Close()

// currently available readers in the pool
current := p.Len()
```


## Credits

 * [Moises P. Sena](https://github.com/moisespsena)
 * [Fatih Arslan](https://github.com/fatih)
 * [sougou](https://github.com/sougou)

## License

The MIT License (MIT) - see LICENSE for more details
