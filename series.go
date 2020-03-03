package uglycharts

import (
	"errors"
	"math"
)

type Series interface {
	Add(p Point)
	Get(index int) (Point, error)
	Clear()
	ClearFree(capacity int)
	Size() int
	isInvalidated() bool
	setValidated()
	GetPoints() []Point
	setChart(ch Chart)
	setLastIndex()
	getLastIndex() int
}

type floatSeries struct {
	ch          Chart
	invalidated bool
	label       string
	points      []Point
	lastIndex   int
}

// capacity is the initial size of the backing slice
func NewFloatSeries(capacity uint) *floatSeries {
	var s floatSeries
	s.points = make([]Point, 0, capacity)
	s.invalidated = false
	return &s
}

// adds new point and immediately updates chart
func (s *floatSeries) Add(p Point) {
	if s.ch.GetAutoRangingY() {
		y := p.GetY()
		if y < s.ch.GetMinY() {
			s.ch.SetMaxY(math.Floor(y - math.Abs(y*0.05)))
			s.ch.setFullRedraw(true)
		}
		if y > s.ch.GetMaxY() {
			s.ch.SetMaxY(math.Ceil(y + math.Abs(y*0.05)))
			s.ch.setFullRedraw(true)
		}
	}
	if s.ch.GetAutoRangingX() {
		x := p.GetX()
		if x < s.ch.GetMinX() {
			s.ch.SetMaxX(math.Floor(x - math.Abs(x*0.05)))
			s.ch.setFullRedraw(true)
		}
		if x > s.ch.GetMaxX() {
			s.ch.SetMaxX(math.Ceil(x + math.Abs(x*0.05)))
			s.ch.setFullRedraw(true)
		}
	}
	s.points = append(s.points, p)
	s.invalidated = true
	s.ch.GetDrawingArea().QueueDraw()
}

func (s *floatSeries) Get(index int) (Point, error) {
	if index > len(s.points) || index < 0 {
		return nil, errors.New("index out of bounds")
	}
	return s.points[index], nil
}

// reslices backing slices without shrinking the underlying array
func (s *floatSeries) Clear() {
	s.points = s.points[:0]
	s.invalidated = true
	s.ch.setFullRedraw(true)
	s.ch.GetDrawingArea().QueueDraw()
}

// recreates backing slices with 0 capacity
func (s *floatSeries) ClearFree(capacity int) {
	s.points = make([]Point, 0, capacity)
	s.invalidated = true
	s.ch.setFullRedraw(true)
	s.ch.GetDrawingArea().QueueDraw()
}

// store reference to the corresponding chart
func (s *floatSeries) setChart(ch Chart) {
	s.ch = ch
}

func (s *floatSeries) GetPoints() []Point {
	return s.points
}

func (s *floatSeries) Size() int {
	return len(s.points)
}

func (s *floatSeries) setLastIndex() {
	s.lastIndex = len(s.points) - 1
	if s.lastIndex == -1 {
		s.lastIndex = 0
	}
}

func (s *floatSeries) getLastIndex() int {
	return s.lastIndex
}

func (s *floatSeries) isInvalidated() bool {
	return s.invalidated
}

func (s *floatSeries) setValidated() {
	s.invalidated = false
}
