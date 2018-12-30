package ziptools

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"sync"
)

// Create a Pool that contains previously used Writers and
// can create new ones if we run out.
var zipWriters = sync.Pool{New: func() interface{} {
	//May want to set a level in the future NewWriterLevel
	var buf bytes.Buffer
	return gzip.NewWriter(&buf)
}}

var zipReaders = sync.Pool{New: func() interface{} {
	// A reader fails if not initialized with some zip data so creating some dummy data when initializing
	var buf bytes.Buffer
	data := []byte{}
	_ = getGzipData(&buf, data)
	rdr, err := gzip.NewReader(&buf)
	//log.Printf("At this point, rdr: %v", rdr)
	if err != nil {
		log.Printf("I am failing because..: %v", err)
	}
	return rdr
}}

func GetGzipData(writer io.Writer, data []byte) error {
	gz := zipWriters.Get().(*gzip.Writer)
	defer zipWriters.Put(gz)
	defer gz.Close()

	// Reset b/c is stateful otherwise
	gz.Reset(writer)
	_, err := gz.Write(data)
	if err != nil {
		return err
	}
	err = gz.Flush()
	if err != nil {
		return err
	}
	return nil
}

func GetGunzipData(writer io.Writer, data []byte) error {
	gz := zipReaders.Get().(*gzip.Reader)
	defer zipReaders.Put(gz)
	defer gz.Close()

	//log.Printf("Possible?  Gzip is: %v", gz)
	err := gz.Reset(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	unzippeddata, err := ioutil.ReadAll(gz)
	if err != nil {
		return err
	}
	writer.Write(unzippeddata)
	return nil
}
