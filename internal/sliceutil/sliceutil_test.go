package sliceutil

import "testing"

// assertReset は Reset の性質を型 T について確かめる
func assertReset[T any](t *testing.T) {
	t.Helper()

	t.Run("nil slice returns nil", func(t *testing.T) {
		got := Reset[T](nil)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("small capacity preserves backing array", func(t *testing.T) {
		s := make([]T, 10, shrinkThreshold)
		got := Reset(s)
		if len(got) != 0 {
			t.Errorf("expected len 0, got %d", len(got))
		}
		if cap(got) != shrinkThreshold {
			t.Errorf("expected cap %d, got %d", shrinkThreshold, cap(got))
		}
	})

	t.Run("large capacity returns nil to release memory", func(t *testing.T) {
		s := make([]T, 0, shrinkThreshold+1)
		got := Reset(s)
		if got != nil {
			t.Errorf("expected nil, got slice with cap %d", cap(got))
		}
	})
}

func TestReset(t *testing.T) {
	// iterator は [][]byte、output は []string で使うので両方まわす
	t.Run("[][]byte", assertReset[[]byte])
	t.Run("[]string", assertReset[string])
}
