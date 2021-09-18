package column

import (
	"errors"
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

func (i IndexSelector) Select(strings []string) ([]string, error) {
	if len(strings) < i.index {
		return nil, errors.New("index out of range")
	}
	if i.index == 0 {
		return strings, nil
	}

	return []string{strings[i.index]}, nil
}
