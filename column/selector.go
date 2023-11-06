package column

import "github.com/xztaityozx/sel/iterator"

type Selector interface {
	Select(w *Writer, iterator iterator.IEnumerable) error
}
