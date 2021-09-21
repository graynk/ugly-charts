package uglycharts

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"image/color"
	"math"
	"strconv"
)

type Chart interface {
	AddSeries(s Series)
	GetDrawingArea() *gtk.DrawingArea
	toLocalCoordinates(p Point) Point
	toRealCoordinates(p Point) Point
	SetTitle(title string)
	ShowLegend(show bool)
	SetAutoRangingX(value bool)
	GetAutoRangingX() bool
	SetAutoRangingY(value bool)
	GetAutoRangingY() bool
	SetDrawMarker(value bool)
	SetColorScheme(value color.Palette)
	SetMarkerSize(value float64)
	SetLineWidth(value float64)
	SetMinX(value float64)
	GetMinX() float64
	SetMaxX(value float64)
	GetMaxX() float64
	SetMinY(value float64)
	GetMinY() float64
	SetMaxY(value float64)
	GetMaxY() float64
	setFullRedraw(value bool)
}

type lineChart struct {
	defaultAxisColor           color.Color
	da                         *gtk.DrawingArea
	autoRangingX, autoRangingY bool
	xLabel, yLabel             string
	fullRedraw                 bool
	drawMarker                 bool
	a, aReal                   Point
	colorScheme                color.Palette
	markerSize                 float64
	lineWidth                  float64
	minX, maxX                 float64
	minY, maxY                 float64
	leftPadding, bottomPadding float64
	series                     []Series
	rightPadding, topPadding   float64
	title                      string
	showLegend                 bool
}

func cairoRGBA(c color.Color) (float64, float64, float64, float64) {
	r, g, b, a := c.RGBA()
	// RGBA() returns values in range [0..0xFFFF] and cairo accepts values in range [0..1]
	return float64(r) / 0xFFFF, float64(g) / 0xFFFF, float64(b) / 0xFFFF, float64(a) / 0xFFFF
}

// NewLineChart
// Constructor that initializes backing slice and connects
// drawing function to GtkDrawingArea
func NewLineChart(da *gtk.DrawingArea) Chart {
	defaultColorScheme := color.Palette{
		color.RGBA{R: 0xF3, G: 0x62, B: 0x2D, A: 0xFF},
		color.RGBA{R: 0xFB, G: 0xA7, B: 0x1B, A: 0xFF},
		color.RGBA{R: 0x57, G: 0xB7, B: 0x57, A: 0xFF},
		color.RGBA{R: 0x41, G: 0xA9, B: 0xC9, A: 0xFF},
		color.RGBA{R: 0x42, G: 0x58, B: 0xC9, A: 0xFF},
		color.RGBA{R: 0x9A, G: 0x42, B: 0xC8, A: 0xFF},
		color.RGBA{R: 0xC8, G: 0x41, B: 0x64, A: 0xFF},
		color.RGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF},
	}
	var ch lineChart
	ch.defaultAxisColor = color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xAA}
	ch.colorScheme = defaultColorScheme
	ch.markerSize = 3
	ch.autoRangingX = true
	ch.autoRangingY = true
	ch.fullRedraw = true
	ch.maxX = 100
	ch.maxY = 100
	ch.da = da
	ch.series = make([]Series, 0)
	ch.a = &FloatPoint{0, 0}
	var backingImage *cairo.Surface
	var icr *cairo.Context
	_ = ch.da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		intWidth, intHeight := da.GetAllocatedWidth(), da.GetAllocatedHeight()
		// not sure how much sense drawing to an image off screen makes, but I'm guessing it's faster with locked axis
		// needs profiling to be sure
		// get DPI which we use to guess appropriate axes padding and font size
		//TODO: just save dpi and w/h
		screen, err := da.GetScreen()
		if err != nil {
			panic(err)
		}
		dpi := screen.GetResolution()
		// axis should be drawn some distance away from the edge of the chart
		width := float64(intWidth)
		height := float64(intHeight)
		if ch.fullRedraw || backingImage == nil || (backingImage.GetWidth() != intWidth || backingImage.GetHeight() != intHeight) {
			ch.bottomPadding = height
			ch.rightPadding = 20
			if ch.title != "" {
				cr.SetFontSize(dpi / 4)
				extents := cr.TextExtents(ch.title)
				ch.topPadding = extents.Height + 20
			} else {
				ch.topPadding = 20
			}
			ch.fullRedraw = true
			backingImage = cairo.CreateImageSurface(cairo.FORMAT_ARGB32, intWidth, intHeight)
			icr = cairo.Create(backingImage)
			// draw axes
			ch.drawAxes(icr, width, height, dpi)
		}
		index := 0
		for i, s := range ch.series {
			if s.isInvalidated() || ch.fullRedraw {
				if !ch.fullRedraw {
					index = s.getLastIndex()
				}
				ch.drawSeriesFromIndex(s, ch.colorScheme[i], index, icr)
				s.setValidated()
				s.setLastIndex()
			}
		}
		if ch.fullRedraw {
			ch.drawTitle(icr, width, dpi)
		}
		ch.fullRedraw = false
		cr.SetSourceSurface(backingImage, 0, 0)
		cr.Paint()
	})
	return &ch
}

func (ch *lineChart) drawAxes(icr *cairo.Context, width, height, dpi float64) {
	r, g, b, a := cairoRGBA(ch.defaultAxisColor)
	icr.SetSourceRGBA(r, g, b, a)
	icr.SetFontSize(math.Floor(dpi / 6)) // arbitrary magic to choose font size
	// TODO: negative numbers take more symbols and it's a mess right now
	maxXLabel := strconv.FormatFloat(ch.maxX, 'f', -1, 64)
	maxYLabel := strconv.FormatFloat(ch.maxY, 'f', -1, 64)
	textXExtents := icr.TextExtents(maxXLabel)
	textYExtents := icr.TextExtents(maxYLabel)
	ch.leftPadding = textYExtents.Width + 20
	// it's not actually bottom padding itself, rather height-padding, resulting in final height at which to draw
	ch.bottomPadding -= textXExtents.Height + 20
	// it's equation of a straight line by 2 dots, where on one axis are local coords and on the other real ones
	ch.a = FloatPoint{
		(width - ch.leftPadding - ch.rightPadding) / (ch.maxX - ch.minX),
		(ch.topPadding - ch.bottomPadding) / (ch.maxY - ch.minY)}
	leftPadding := ch.leftPadding
	bottomPadding := ch.bottomPadding
	topPadding := ch.topPadding
	rightPadding := ch.rightPadding
	zero := ch.toLocalCoordinates(FloatPoint{0, 0})
	min := ch.toLocalCoordinates(FloatPoint{ch.minX, ch.minY})
	max := ch.toLocalCoordinates(FloatPoint{ch.maxX, ch.maxY})
	icr.MoveTo(zero.GetX(), max.GetY())
	icr.LineTo(zero.GetX(), min.GetY()+5)
	icr.MoveTo(min.GetX(), zero.GetY())
	icr.LineTo(max.GetX(), zero.GetY())
	icr.Stroke()
	axisWidth := width - leftPadding - rightPadding
	count := axisWidth * 0.25 / textXExtents.Width // more arbitrary magic to figure out how many labels can fit
	if count < 2 {
		count = 2
	}
	deltaX := ch.maxX - ch.minX
	// we default to integers markers, but if delta is low enough we start using decimals
	var stepX float64
	var precision int
	stepX = math.Ceil(deltaX / count)
	//if deltaX > 10 {
	//	stepX = math.Ceil(deltaX / count)
	//	precision = -1
	//} else {
	//	stepX = deltaX / count
	//	precision = 2
	//}
	yPosition := zero.GetY() + textXExtents.Height + dpi/12 // you know the drill
	size := icr.TextExtents("0")
	halfWidth := size.Width / 2
	icr.MoveTo(zero.GetX()-halfWidth, yPosition)
	icr.ShowText("0")
	for i := float64(1); i < count; i++ {
		label := strconv.FormatFloat(i*stepX+ch.minX, 'f', precision, 64)
		xPosition := ch.toLocalCoordinates(FloatPoint{i*stepX + ch.minX, 0}).GetX()
		icr.MoveTo(xPosition, zero.GetY()+5)
		icr.LineTo(xPosition, zero.GetY()-5)
		icr.Stroke()
		icr.SetSourceRGBA(r, g, b, a/3)
		icr.SetDash([]float64{2, 2}, 0)
		icr.MoveTo(xPosition, bottomPadding-10)
		icr.LineTo(xPosition, topPadding)
		icr.Stroke()
		icr.SetDash([]float64{}, 0)
		icr.SetSourceRGBA(r, g, b, a)
		size := icr.TextExtents(label)
		halfWidth := size.Width / 2
		icr.MoveTo(xPosition-halfWidth, yPosition)
		icr.ShowText(label)
	}

	// same for Y axis
	axisHeight := bottomPadding - topPadding
	count = axisHeight * 0.25 / textYExtents.Height
	if count < 2 {
		count = 2
	}
	deltaY := ch.maxY - ch.minY
	var stepY float64
	stepY = math.Ceil(deltaY / count)
	//if deltaY > 10 {
	//	stepY = math.Ceil(deltaY / count)
	//	precision = -1
	//} else {
	//	stepY = deltaY / count
	//	precision = 2
	//}
	for i := float64(1); i < count+1; i++ {
		yPosition := ch.toLocalCoordinates(FloatPoint{0, i*stepY + ch.minY}).GetY()
		if yPosition < ch.topPadding {
			break
		}
		icr.MoveTo(zero.GetX()-5, yPosition)
		icr.LineTo(zero.GetX()+5, yPosition)
		icr.Stroke()
		icr.SetSourceRGBA(r, g, b, a/3)
		icr.SetDash([]float64{2, 2}, 0)
		icr.MoveTo(leftPadding+10, yPosition)
		icr.LineTo(width-rightPadding, yPosition)
		icr.Stroke()
		icr.SetDash([]float64{}, 0)
		icr.SetSourceRGBA(r, g, b, a)
		label := strconv.FormatFloat(i*stepY+ch.minY, 'f', precision, 64)
		// have to calculate it each time, 'cause height is eternal, but width is everchanging
		size := icr.TextExtents(label)
		xPosition := zero.GetX() - size.Width - 10 // you know the drill
		icr.MoveTo(xPosition, yPosition+size.Height/2)
		icr.ShowText(label)
	}
}

func (ch *lineChart) drawTitle(icr *cairo.Context, width, dpi float64) {
	if ch.title == "" {
		return
	}
	icr.SetFontSize(dpi / 4)
	extents := icr.TextExtents(ch.title)
	x, y := (width-extents.Width)/2, extents.Height+5
	icr.SetSourceRGBA(0.3, 0.3, 0.3, 0.2)
	icr.Rectangle(x-5, y-extents.Height, extents.Width+15, extents.Height+10)
	icr.Fill()
	icr.SetSourceRGBA(0, 0, 0, 1)
	icr.MoveTo(x, y)
	icr.ShowText(ch.title)
}

func (ch *lineChart) drawSeriesFromIndex(s Series, scheme color.Color, index int, icr *cairo.Context) {
	// set appropriate color for the series
	icr.SetSourceRGBA(cairoRGBA(scheme))
	icr.SetLineWidth(ch.lineWidth)
	// please don't change this concurrently, I guess
	points := s.GetPoints()
	pointsLength := len(points)
	if pointsLength == 0 {
		return
	}
	// calculate local points
	localPoints := make([]Point, pointsLength-index)
	for i := range localPoints {
		localPoints[i] = ch.toLocalCoordinates(points[index+i])
	}
	// first draw all the lines in the series
	point := localPoints[0]
	icr.MoveTo(point.GetX(), point.GetY())
	for i := 1; i < len(localPoints); i++ {
		point = localPoints[i]
		icr.LineTo(point.GetX(), point.GetY())
	}
	icr.Stroke()
	if ch.drawMarker {
		for _, point := range localPoints {
			icr.Arc(point.GetX(), point.GetY(), ch.markerSize, 0, 2*math.Pi)
			icr.Fill()
		}
	}
}

// it's equation of a straight line by 2 dots, where on one axis are local coords and on the other real ones
func (ch *lineChart) toLocalCoordinates(value Point) Point {
	x, y := value.GetCoordinates()
	aX, aY := ch.a.GetCoordinates()
	minX, minY := ch.minX, ch.minY
	bX, bY := ch.leftPadding, ch.bottomPadding
	xLocal := pointOnLine(x, aX, minX, bX)
	yLocal := pointOnLine(y, aY, minY, bY)
	return FloatPoint{xLocal, yLocal}
}

// converting to real means just swapping ranges around and inverting "a" coefficient
func (ch *lineChart) toRealCoordinates(value Point) Point {
	x, y := value.GetCoordinates()
	aX, aY := ch.a.GetCoordinates()
	minX, minY := ch.leftPadding, ch.bottomPadding
	bX, bY := ch.minX, ch.minY
	xLocal := pointOnLine(x, 1/aX, minX, bX)
	yLocal := pointOnLine(y, 1/aY, minY, bY)
	return FloatPoint{xLocal, yLocal}
}

func pointOnLine(value, a, min, b float64) float64 {
	return (value-min)*a + b
}

func (ch *lineChart) AddSeries(s Series) {
	ch.series = append(ch.series, s)
	s.setChart(ch)
	if len(s.GetPoints()) != 0 {
		ch.da.QueueDraw()
	}
}

func (ch *lineChart) GetDrawingArea() *gtk.DrawingArea {
	return ch.da
}

func (ch *lineChart) SetAutoRangingX(value bool) {
	ch.autoRangingX = value
}

func (ch *lineChart) GetAutoRangingX() bool {
	return ch.autoRangingX
}

func (ch *lineChart) SetAutoRangingY(value bool) {
	ch.autoRangingY = value
}

func (ch *lineChart) GetAutoRangingY() bool {
	return ch.autoRangingY
}

func (ch *lineChart) SetDrawMarker(value bool) {
	ch.drawMarker = value
}

func (ch *lineChart) SetColorScheme(value color.Palette) {
	ch.colorScheme = value
}

func (ch *lineChart) SetMarkerSize(value float64) {
	ch.markerSize = value
}

func (ch *lineChart) SetLineWidth(value float64) {
	ch.lineWidth = value
}

func (ch *lineChart) SetMinX(value float64) {
	ch.minX = value
}

func (ch *lineChart) GetMinX() float64 {
	return ch.minX
}

func (ch *lineChart) SetMaxX(value float64) {
	ch.maxX = value
}

func (ch *lineChart) GetMaxX() float64 {
	return ch.maxX
}

func (ch *lineChart) SetMinY(value float64) {
	ch.minY = value
}

func (ch *lineChart) GetMinY() float64 {
	return ch.minY
}

func (ch *lineChart) SetMaxY(value float64) {
	ch.maxY = value
}

func (ch *lineChart) GetMaxY() float64 {
	return ch.maxY
}

func (ch *lineChart) setFullRedraw(value bool) {
	ch.fullRedraw = value
}

func (ch *lineChart) SetTitle(title string) {
	ch.fullRedraw = true
	ch.title = title
	ch.da.QueueDraw()
}

func (ch *lineChart) ShowLegend(show bool) {
	ch.showLegend = show
}
