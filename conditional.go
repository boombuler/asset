package asset

// Creates a conditional pipeliner.
// It calls the condition for every asset and calls
// the given pipeliners only for those assets, for which
// the condition evaluates true.
// For example: only compress JS when in production mode.
func If(condition func(Asset) bool, pipes ...Pipeliner) Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)

		pipeIn := make(chan Asset, 1)
		go func(in <-chan Asset) {
			for _, p := range pipes {
				in = p.Pipe(in)
			}

			for a := range in {
				result <- a
			}
			close(result)
		}(pipeIn)

		go func() {
			for asset := range input {
				if condition(asset) {
					pipeIn <- asset
				} else {
					result <- asset
				}
			}
			close(pipeIn)
		}()
		return result
	})
}
