// Package sliceutil は使い回すスライスのためのちいさなヘルパを提供する
package sliceutil

// shrinkThreshold はスライスを再利用するかどうかの境目
const shrinkThreshold = 64

// Reset はスライスを長さ0にリセットする。
// 容量が shrinkThreshold を超えている場合は nil を返し、backing array を GC 可能にする。
// それ以外の場合は [:0] で容量を維持して再利用する
func Reset[T any](s []T) []T {
	if cap(s) > shrinkThreshold {
		return nil
	}
	return s[:0]
}
