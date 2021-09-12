package gen

import "github.com/cheekybits/genny/generic"

type TSource generic.Type
type TResult generic.Type

type TSourceSlice []TSource
type TResultSlice []TResult

func (t TSourceSlice) Where(predicate func(source TSource) bool) (result TSourceSlice) {
	for _, value := range t {
		if predicate(value) {
			result = append(result, value)
		}
	}
	return
}

func (t TSourceSlice) Select(selector func(source TSource) TResult) (result TResultSlice) {
	for _, value := range t {
		result = append(result, selector(value))
	}
	return
}
