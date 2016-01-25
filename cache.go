package asset

import (
	"bytes"
	"io"
	"io/ioutil"
)

type cachedAsset struct {
	data []byte
	name string
}

func (ca *cachedAsset) GetName() string {
	return ca.name
}

func (ca *cachedAsset) GetContent() (io.ReadCloser, error) {
	buf := bytes.NewBuffer(ca.data)
	return ioutil.NopCloser(buf), nil
}

// Creates a Cache-Pipeline which keeps the resulting content in memory.
// Should be used as the last content transforming pipeline...
func Cache() Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)
		go func() {
			for a := range input {
				name := a.GetName()
				content, err := a.GetContent()
				if err == nil {
					data, _ := ioutil.ReadAll(content)
					result <- &cachedAsset{data, name}
				}
				if content != nil {
					content.Close()
				}
			}
			close(result)
		}()
		return result
	})
}
