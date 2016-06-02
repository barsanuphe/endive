package book

type Books []Book

// filter Books with a given function
func (bks Books) filter(f func(*Book) bool) (filteredBooks Books) {
	filteredBooks = make(Books, 0)
	for _, v := range bks {
		if f(&v) {
			filteredBooks = append(filteredBooks, v)
		}
	}
	return
}

// FilterIncomplete among Books.
func (bks *Books) FilterIncomplete() Books {
	return bks.filter(func(b *Book) bool { return !b.Metadata.IsComplete() })
}

// FilterByProgress among Books.
func (bks *Books) FilterByProgress(progress string) Books {
	return bks.filter(func (b *Book) bool { return b.Progress == progress })
}

// FilterUntagged among Books.
func (bks *Books) FilterUntagged() Books {
	return bks.filter(func (b *Book) bool { return len(b.Metadata.Tags) == 0 })
}

// FilterRetail among Books.
func (bks *Books) FilterRetail() Books {
	return bks.filter(func (b *Book) bool { return b.HasRetail() })
}

func (bks *Books) FilterNonRetailOnly() Books {
	return bks.filter(func (b *Book) bool { return !b.HasRetail() })
}
