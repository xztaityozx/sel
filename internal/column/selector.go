package column

import (
	"github.com/xztaityozx/sel/internal/iterator"
)

type Selector interface {
	Select(w *Writer, iterator iterator.IEnumerable) error
}
