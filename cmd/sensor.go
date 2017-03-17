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

import "reflect"

// Detects if one should display camera.
func (v *xrayHandlers) shouldDisplayCamera(sr sensorRecord) bool {
	v.RLock()
	prevSR := v.prevSR
	v.RUnlock()
	return !reflect.DeepEqual(prevSR.Values, sr.Values)
}

type sensorRecord struct {
	Name      string      `json:"sensorName"`
	Type      int         `json:"sensorType"`
	Timestamp int         `json:"timestamp"`
	Accuracy  int         `json:"accuracy"`
	Values    [][]float64 `json:"values"`
}

// Saves current sensor data for motion detection.
func (v *xrayHandlers) persistCurrentSensorR(sr sensorRecord) {
	v.Lock()
	v.prevSR = sr
	v.Unlock()
}
