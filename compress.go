package compress

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"io/ioutil"
	"sync"
)

// Proper usage of a sync.Pool requires each entry to have approximately
// the same memory cost. To obtain this property when the stored type
// contains a variably-sized buffer, we add a hard limit on the maximum buffer
// to place back in the pool.
//
// See https://golang.org/issue/23199
//if cap(p.buf) > 64<<10 {
//	return
//}

// 压缩buffer对象池
var CompressBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// 压缩器对象池
var CompressWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := flate.NewWriter(nil, flate.BestSpeed)
		return w
	},
}

// 解压器对象池
var DecompressWriterPool = sync.Pool{
	New: func() interface{} {
		r := flate.NewReader(nil)
		return r
	},
}

// Compress 压缩
func Compress(data []byte) (ret []byte, err error) {
	writer := CompressWriterPool.Get().(*flate.Writer)
	buffer := CompressBufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	writer.Reset(buffer)
	if _, err = writer.Write(data); err != nil {
		if err1 := writer.Close(); err1 != nil {
			err = errors.New(err.Error() + " " + err1.Error())
		}
		return
	}
	if err = writer.Flush(); err != nil {
		if err1 := writer.Close(); err1 != nil {
			err = errors.New(err.Error() + " " + err1.Error())
		}
		return
	}
	if err = writer.Close(); err != nil {
		return
	}
	ret, err = ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}
	//ret = buffer.Bytes()
	//buffer.Reset()
	CompressBufferPool.Put(buffer)
	CompressWriterPool.Put(writer)
	return
}

// Decompress 解压缩
func Decompress(data []byte) ([]byte, error) {
	reader := DecompressWriterPool.Get().(io.ReadCloser)
	buffer := CompressBufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	buffer.Write(data)
	// Reset the decompressor and decode to some output stream.
	if err := reader.(flate.Resetter).Reset(buffer, nil); err != nil {
		return nil, err
	}
	defer reader.Close()
	out, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	//buffer.Reset()
	CompressBufferPool.Put(buffer)
	DecompressWriterPool.Put(reader)
	return out, nil
}
