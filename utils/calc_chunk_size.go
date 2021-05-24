package utils

import "math"

func CalculateChunkSize(size int64) int64 {
	sizeF := float64(size)
	// twitter api says that chunks cannot exceed >=5MB, so regardless of file size,
	// 4 chunks will always be sent
	return int64(math.Floor(sizeF * 0.30))
}
