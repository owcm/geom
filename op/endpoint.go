// Copyright (c) 2011 Mateusz Czapliński (Go port)
// Copyright (c) 2011 Mahir Iqbal (as3 version)
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// based on http://code.google.com/p/as3polyclip/ (MIT licensed)
// and code by Martínez et al: http://wwwdi.ujaen.es/~fmartin/bool_op.html (public domain)

package op

import (
	"fmt"

	"github.com/geodatalake/geom"
	"github.com/gonum/floats"
)

// A container for endpoint data. A endpoint represents a location of interest (vertex between two polygon edges)
// as the sweep line passes through the polygons.
type endpoint struct {
	p           geom.Point
	left        bool      // Is the point the left endpoint of the segment (p, other->p)?
	polygonType           // polygonType to which this event belongs to
	other       *endpoint // Event associated to the other endpoint of the segment

	// Does the segment (p, other->p) represent an inside-outside transition
	// in the polygon for a vertical ray from (p.x, -infinite) that crosses the segment?
	inout bool
	edgeType
	inside bool // Only used in "left" events. Is the segment (p, other->p) inside the other polygon?
}

func (e endpoint) String() string {
	sleft := map[bool]string{true: "left", false: "right"}
	return fmt.Sprint("{", e.p, " ", sleft[e.left], " type:", e.polygonType,
		" other:", e.other.p, " inout:", e.inout, " inside:", e.inside, " edgeType:", e.edgeType, "}")
}

func (e1 *endpoint) equals(e2 *endpoint) bool {
	return PointEquals(e1.p, e2.p) &&
		e1.left == e2.left &&
		e1.polygonType == e2.polygonType &&
		e1.inout == e2.inout &&
		e1.edgeType == e2.edgeType &&
		e1.inside == e2.inside &&
		PointEquals(e1.other.p, e2.other.p) &&
		e1.other.left == e2.other.left &&
		e1.other.polygonType == e2.other.polygonType &&
		e1.other.inout == e2.other.inout &&
		e1.other.edgeType == e2.other.edgeType &&
		e1.other.inside == e2.other.inside
}

func (se *endpoint) segment() segment {
	return segment{se.p, se.other.p}
}

func signedArea(p0, p1, p2 geom.Point) float64 {
	return (p0.X-p2.X)*(p1.Y-p2.Y) -
		(p1.X-p2.X)*(p0.Y-p2.Y)
}

// Checks if this sweep event is below point p.
func (se *endpoint) below(x geom.Point) bool {
	if se.left {
		a := signedArea(se.p, se.other.p, x)
		return !floats.EqualWithinULP(a, 0, 2) && a > 0
	}
	a := signedArea(se.other.p, se.p, x)
	return !floats.EqualWithinULP(a, 0, 2) && a > 0
}

func (se *endpoint) above(x geom.Point) bool {
	return !se.below(x)
}
