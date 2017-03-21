/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"image"
	"math"
	"time"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// XorRects returns "XOR" difference between two rectangles
func XorRects(r, s image.Rectangle) []image.Rectangle {

	if r.Intersect(s).Empty() {
		// In case there is no overlap at all, return both rectangles
		return []image.Rectangle{r, s}
	}

	// Else determine difference
	a := min(r.Min.X, s.Min.X)
	b := max(r.Min.X, s.Min.X)
	c := min(r.Max.X, s.Max.X)
	d := max(r.Max.X, s.Max.X)

	e := min(r.Min.Y, s.Min.Y)
	f := max(r.Min.Y, s.Min.Y)
	g := min(r.Max.Y, s.Max.Y)
	h := max(r.Max.Y, s.Max.Y)

	// X = intersection, 0-7 = possible difference areas
	// h +-+-+-+
	// . |5|6|7|
	// g +-+-+-+
	// . |3|X|4|
	// f +-+-+-+
	// . |0|1|2|
	// e +-+-+-+
	// . a b c d

	result := []image.Rectangle{}

	// we'll always have rectangles 1, 3, 4 and 6
	result = append(result, image.Rect(b, e, c, f))
	result = append(result, image.Rect(a, f, b, g))
	result = append(result, image.Rect(c, f, d, g))
	result = append(result, image.Rect(b, g, c, h))

	// decide which corners
	if r.Min.X == a && r.Min.Y == e || s.Min.X == a && s.Min.Y == e {
		// corners 0 and 7
		result = append(result, image.Rect(a, e, b, f))
		result = append(result, image.Rect(c, g, d, h))
	} else {
		// corners 2 and 5
		result = append(result, image.Rect(c, e, d, f))
		result = append(result, image.Rect(a, g, b, h))
	}

	return result
}

type motionRecorder struct {
	prevFrame          *frameRecord
	frameMotions       []float64
	snapshotTimestamps []time.Time
}

func findClosestRectangle(prevFace image.Rectangle, nextFaces []image.Rectangle) int {

	prevCenterX := prevFace.Min.X + prevFace.Dx()
	prevCenterY := prevFace.Min.Y + prevFace.Dy()

	n, distance := -1, math.MaxInt64
	for j := 0; j < len(nextFaces); j++ {
		nextCenterX := nextFaces[j].Min.X + nextFaces[j].Dx()
		nextCenterY := nextFaces[j].Min.Y + nextFaces[j].Dy()

		di := (prevCenterX-nextCenterX)*(prevCenterX-nextCenterX) + (prevCenterY-nextCenterY)*(prevCenterY-nextCenterY)
		if di < distance {
			distance = di
			n = j
		}
	}

	return n
}

func sumAreas(rects []image.Rectangle) float64 {

	diff := 0
	for _, r := range rects {
		diff += r.Dx() * r.Dy()
	}

	return float64(diff)
}

func analyseBetweenFrames(prev, next *frameRecord) float64 {

	prevFaces, _ := prev.GetFaceRectangles()
	nextFaces, _ := next.GetFaceRectangles()

	prevLen := len(prevFaces)
	nextLen := len(nextFaces)

	l := min(prevLen, nextLen)

	result := float64(0.0)

	i := 0
	// Handle faces present in both frame
	for ; i < l; i++ {

		n := findClosestRectangle(prevFaces[i], nextFaces)

		xorRects := XorRects(prevFaces[i], nextFaces[n])

		result += sumAreas(xorRects)
	}
	// Handle faces only available in previous frame
	for p := i; p < len(prevFaces); p++ {

		xorRects := XorRects(prevFaces[p], image.Rectangle{})

		result += sumAreas(xorRects)
	}
	// Handle faces only available in next frame
	for n := i; n < len(nextFaces); n++ {

		xorRects := XorRects(image.Rectangle{}, nextFaces[n])

		result += sumAreas(xorRects)
	}

	frame, _, _ := prev.GetFullFrameRect()

	// Normalize by pixels for screen size
	return result / float64(frame.Dx()*frame.Dy())
}

func (mr *motionRecorder) analyze() float64 {

	result := float64(0.0)

	for _, motion := range mr.frameMotions {
		result += motion
	}

	return result
}

const maxFrames = 1000 // Maximum number of frames to keep track off

const thresholdBase = 0.05  // Base value for threshold
const thresholdBoost = 0.10 // Maximum boost for threshold when timestamps are taken

const maxTimestamps = 10                     // Maximum number of timestamps to keep
const minimalTimestampDiff = time.Second * 5 // Minimal difference between timestamps
const maxTimestampsAge = time.Second * 30    // Age to remove recorded timestamp from array

func (mr *motionRecorder) Threshold() float64 {

	// Detect older time stamps
	itime := len(mr.snapshotTimestamps) - 1
	for ; itime >= 0; itime-- {
		if time.Since(mr.snapshotTimestamps[itime]) >= maxTimestampsAge {
			break
		}
	} // and remove them
	if itime > 0 {
		mr.snapshotTimestamps = mr.snapshotTimestamps[itime:]
	}

	return thresholdBase + thresholdBoost*float64(len(mr.snapshotTimestamps))/maxTimestamps
}

func (mr *motionRecorder) Append(fr *frameRecord) {

	if mr.prevFrame != nil {

		diff := analyseBetweenFrames(mr.prevFrame, fr)

		mr.frameMotions = append(mr.frameMotions, diff)
		if len(mr.frameMotions) > maxFrames {
			mr.frameMotions = mr.frameMotions[1:]
		}
	}

	mr.prevFrame = fr
}

func (mr *motionRecorder) DetectMotion() bool {

	if len(mr.snapshotTimestamps) > 0 {
		if time.Since(mr.snapshotTimestamps[len(mr.snapshotTimestamps)-1]) < minimalTimestampDiff {
			return false
		}
	}

	activity := mr.analyze()
	if activity >= mr.Threshold() {
		mr.snapshotTimestamps = append(mr.snapshotTimestamps, time.Now())
		if len(mr.snapshotTimestamps) > maxTimestamps {
			mr.snapshotTimestamps = mr.snapshotTimestamps[1:]
		}
		return true
	}

	return false
}
