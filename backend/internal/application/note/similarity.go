package note

import "math"

// cosineSimilarity returns the cosine of the angle between two vectors: 1 when
// they point the same way, 0 when orthogonal, -1 when opposed. It returns 0
// for empty or zero-length vectors, and compares only the overlapping prefix
// when the two differ in length.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
		if i >= len(b) {
			break
		}
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / float32(math.Sqrt(float64(normA))*math.Sqrt(float64(normB)))
}
