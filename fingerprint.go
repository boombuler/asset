package asset

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
)

type fingerPrinted struct {
	Asset
	filename string
}

func (fp fingerPrinted) GetName() string {
	return fp.filename
}

// A func which creates a new filename for a fingerprinted file.
// base is the basic filename without extension.
// fp is the fingerprint of the file.
// ext is the extension of the file.
type FingerprintFunc func(base, fp, ext string) string

// Default fingerprint function which returns a new filename like: base_hash.js
var FingerprintDefault = FingerprintFunc(func(base, fp, ext string) string {
	return fmt.Sprintf("%s_%s%s", base, fp, ext)
})

// Fingerprint function which removes the base filename and only returns the hash and extension.
var FingerprintHashOnly = FingerprintFunc(func(base, fp, ext string) string {
	return fmt.Sprintf("%%s%s", fp, ext)
})

// Fingerprint appends a hash of the file to the filename.
// This can be used for cachebusting.
// The given FingerprintFunc will be used to calculate the resulting filename.
func Fingerprint(fpf FingerprintFunc) Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)
		go func() {
			for a := range input {
				hash := md5.New()
				if content, err := a.GetContent(); err == nil {
					io.Copy(hash, content)
					content.Close()
					fp := fmt.Sprintf("%x", hash.Sum(nil))
					base := a.GetName()
					ext := filepath.Ext(base)
					base = base[:len(base)-len(ext)]
					fileName := fpf(base, fp, ext)
					result <- fingerPrinted{a, fileName}
				}
			}
			close(result)
		}()
		return result
	})
}
