package output

import "testing"

func TestResetStringSlice(t *testing.T) {
	t.Run("nil slice returns nil", func(t *testing.T) {
		got := resetStringSlice(nil)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("small capacity preserves backing array", func(t *testing.T) {
		s := make([]string, 10, shrinkThreshold)
		got := resetStringSlice(s)
		if len(got) != 0 {
			t.Errorf("expected len 0, got %d", len(got))
		}
		if cap(got) != shrinkThreshold {
			t.Errorf("expected cap %d, got %d", shrinkThreshold, cap(got))
		}
	})

	t.Run("large capacity returns nil to release memory", func(t *testing.T) {
		s := make([]string, 0, shrinkThreshold+1)
		got := resetStringSlice(s)
		if got != nil {
			t.Errorf("expected nil, got slice with cap %d", cap(got))
		}
	})
}
