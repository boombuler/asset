package asset

import (
	"errors"
	"io"
	"sync"
)

// Returned by Manager GetContent when an asset is not found
var ErrAssetNotFound = errors.New("Asset not found")

// A function which can be used in a template to inclue assets
// It searches the given file patterns and passes those file through
// the pipeline...
type TemplateFunc func(filePatterns ...string) <-chan string

// Pipeliners process a number of assets and may transform them into something else
type Pipeliner interface {
	Pipe(input <-chan Asset) <-chan Asset
}

// PipeFunc is a Pipeliner implementation based on a func.
type PipeFunc func(input <-chan Asset) <-chan Asset

// Pipes the assets through the Pipefunc.
func (pf PipeFunc) Pipe(input <-chan Asset) <-chan Asset {
	return pf(input)
}

// Some kind of asset
type Asset interface {
	// GetName returns the filename of the asset
	GetName() string
	// GetContents gets a reader for the file content
	GetContent() (io.ReadCloser, error)
}

type getContentFunc func() (io.ReadCloser, error)

// Manager keeps track of the assets which are requested by the TemplateFunc and the Pipeline
type Manager struct {
	m            *sync.Mutex
	files        map[string]getContentFunc
	TemplateFunc TemplateFunc
}

// GetContent returns the content of a given asset
func (am *Manager) GetContent(name string) (io.ReadCloser, error) {
	am.m.Lock()
	defer am.m.Unlock()
	if fn, ok := am.files[name]; ok {
		return fn()
	} else {
		return nil, ErrAssetNotFound
	}
}

// Creates a new Manager which keeps track of assets.
// the given pipelines will be used to transform the assets
func New(pipelines ...Pipeliner) *Manager {
	result := new(Manager)
	result.m = new(sync.Mutex)
	result.files = make(map[string]getContentFunc)
	result.TemplateFunc = func(assets ...string) <-chan string {

		files := resolveFiles(assets)
		for _, p := range pipelines {
			files = p.Pipe(files)
		}
		fileNames := make(chan string)
		go func() {
			result.m.Lock()
			defer result.m.Unlock()
			for asset := range files {
				curAsset := asset
				name := curAsset.GetName()
				result.files[name] = curAsset.GetContent
				fileNames <- name
			}
			close(fileNames)
		}()

		return fileNames
	}
	return result
}
