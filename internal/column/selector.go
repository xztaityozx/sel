package column

import (
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
)

type Selector interface {
	Select(w *output.Writer, columns iterator.Columns) error
}
