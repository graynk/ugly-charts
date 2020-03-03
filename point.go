package uglycharts

import "math"

type Point interface {
	Equals(p Point) bool
	GetX() float64
	GetY() float64
	GetCoordinates() (float64, float64)
}

type IntPoint struct {
	X, Y int
}

type FloatPoint struct {
	X, Y float64
}

func (p IntPoint) GetCoordinates() (float64, float64) {
	return float64(p.X), float64(p.Y)
}

func (p IntPoint) GetX() float64 {
	return float64(p.X)
}

func (p IntPoint) GetY() float64 {
	return float64(p.Y)
}

func (p IntPoint) Equals(other Point) bool {
	return equalFloats(p.GetX(), other.GetX()) && equalFloats(p.GetY(), other.GetY())
}

func (p FloatPoint) GetCoordinates() (float64, float64) {
	return p.X, p.Y
}

func (p FloatPoint) GetX() float64 {
	return p.X
}

func (p FloatPoint) GetY() float64 {
	return p.Y
}

func (p FloatPoint) Equals(other Point) bool {
	return equalFloats(p.GetX(), other.GetX()) && equalFloats(p.GetY(), other.GetY())
}

func equalFloats(a, b float64) bool {
	return math.Abs(a-b) > 1e-10
}
