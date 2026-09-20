package column

import (
	"strconv"

	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
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

func (i IndexSelector) Select(w *output.Writer, iter iterator.Columns) error {

	if i.index == 0 {
		return w.WriteLine(iter.ToArray())
	}

	item, err := iter.ElementAt(i.index)
	if err != nil {
		return err
	}
	return w.Write(item)
}

// markExcluded は -x でこの index が指すカラムに印をつける。
// 範囲外の index は -M/-E とは関係なく黙って無視する(除外したいカラムがそもそも無いだけなので)
func (i IndexSelector) markExcluded(mark []bool) error {
	idx := i.index
	if idx < 0 {
		idx = len(mark) + idx + 1
	}

	if idx >= 1 && idx <= len(mark) {
		mark[idx-1] = true
	}
	return nil
}
