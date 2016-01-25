package asset

import (
	"io"
	"path/filepath"
)

type mergedAsset []Asset

type multiReadCloser struct {
	assets    []Asset
	curReader io.ReadCloser
}

func (mr *multiReadCloser) Read(p []byte) (n int, err error) {
	if mr.curReader == nil && len(mr.assets) > 0 {
		mr.curReader, err = mr.assets[0].GetContent()
		if err != nil {
			return
		}
		mr.assets = mr.assets[1:]
	}

	for mr.curReader != nil {
		n, err = mr.curReader.Read(p)
		if n > 0 || err != io.EOF {
			if err == io.EOF {
				err = nil
			}
			return
		}
		err = mr.curReader.Close()
		mr.curReader = nil

		if err != nil {
			return
		}
		if len(mr.assets) > 0 {
			mr.curReader, err = mr.assets[0].GetContent()
			if err != nil {
				return
			}
			mr.assets = mr.assets[1:]
		}
	}
	return 0, io.EOF
}

func (mr *multiReadCloser) Close() error {
	if mr.curReader != nil {
		return mr.curReader.Close()
	}
	return nil
}

func (m mergedAsset) GetName() string {
	result := ""
	for i, a := range m {
		fn := a.GetName()

		if i > 0 {
			ext := filepath.Ext(fn)
			fn = fn[0 : len(fn)-len(ext)]
			result = fn + "_" + result
		} else {
			result = fn
		}
	}
	return result
}

func (m mergedAsset) GetContent() (io.ReadCloser, error) {
	r := make([]Asset, len(m))
	copy(r, m)
	return &multiReadCloser{r, nil}, nil
}

// Create a merge pipeliner. which combines multiple files into one.
func Merge() Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)
		go func() {
			ma := make(mergedAsset, 0)
			for a := range input {
				ma = append(ma, a)
			}
			result <- ma
			close(result)
		}()
		return result
	})
}
