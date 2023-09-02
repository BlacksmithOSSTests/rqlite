package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
)

const DefaultBufferSize = 65536

// Compressor is a wrapper around a gzip.Writer that reads from an io.Reader
// and writes to an internal buffer.  The internal buffer is used to store
// compressed data until it is read by the caller.
type Compressor struct {
	r     io.Reader
	bufSz int
	buf   *bytes.Buffer
	gzw   *gzip.Writer
}

// NewCompressor returns an instantiated Compressor that reads from r and
// compresses the data using gzip.
func NewCompressor(r io.Reader, bufSz int) *Compressor {
	buf := new(bytes.Buffer)
	return &Compressor{
		r:     r,
		bufSz: bufSz,
		buf:   buf,
		gzw:   gzip.NewWriter(buf),
	}
}

// Read reads compressed data.
func (c *Compressor) Read(p []byte) (int, error) {
	if c.buf.Len() == 0 && c.gzw != nil {
		n, err := io.CopyN(c.gzw, c.r, int64(c.bufSz))
		if err != nil {
			c.gzw.Close() // Time to write the footer.
			if err != io.EOF {
				// Actual error, let caller handle
				return 0, err
			}
			c.gzw = nil
		} else if n > 0 {
			// We read some data, but didn't hit any error.
			// Just flush the data in the buffer, ready
			// to be read.
			if err := c.gzw.Flush(); err != nil {
				return 0, err
			}
		}
	}
	return c.buf.Read(p)
}

// Close closes the Compressor.
func (c *Compressor) Close() error {
	return c.gzw.Close()
}
