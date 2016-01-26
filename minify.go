package asset

import (
	"github.com/tdewolff/minify"
	"io"
	"mime"
	"path/filepath"
)

type minifiedAsset struct {
	base     Asset
	minifier *minify.M
}

type readCloseWrapper struct {
	io.Reader
	io.Closer
}

func (m minifiedAsset) GetName() string {
	return m.base.GetName()
}

func (m minifiedAsset) GetContent() (io.ReadCloser, error) {
	name := m.base.GetName()
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	content, err := m.base.GetContent()
	if err != nil {
		return content, err
	}
	reader := m.minifier.Reader(mimeType, content)
	return &readCloseWrapper{
		Reader: reader,
		Closer: content,
	}, nil
}

func Minify(m *minify.M) Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)
		go func() {
			for a := range input {
				result <- minifiedAsset{
					base:     a,
					minifier: m,
				}
			}
			close(result)
		}()
		return result
	})
}
