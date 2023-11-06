package column

import (
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
	"strconv"
)

// IndexSelector は単一のindexを指定してカラムを選択するやつ
type IndexSelector struct {
	index int
}

func NewIndexSelector(i int) IndexSelector {
	return IndexSelector{index: i}
}

func NewIndexSelectorFromString(str string, def int) (IndexSelector, error) {
	if len(str) == 0 {
		return IndexSelector{index: def}, nil
	}
	num, err := strconv.Atoi(str)
	return NewIndexSelector(num), err
}

func (i IndexSelector) Select(w *output.Writer, iter iterator.IEnumerable) error {

	if i.index == 0 {
		return w.Write(iter.ToArray()...)
	}

	item, err := iter.ElementAt(i.index)
	if err != nil {
		return err
	}
	return w.Write(item)
}
