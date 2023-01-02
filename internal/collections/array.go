package collections

import "sort"

func SortAsc(arr []uint64) {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
}
