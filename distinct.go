package asset

// Pipeliner that accepts each file only once.
func Distinct() Pipeliner {
	return PipeFunc(func(input <-chan Asset) <-chan Asset {
		result := make(chan Asset)
		set := make(map[string]bool)
		go func() {
			for a := range input {
				name := a.GetName()
				if _, found := set[name]; !found {
					set[name] = true
					result <- a
				}
			}
			close(result)
		}()
		return result
	})
}
