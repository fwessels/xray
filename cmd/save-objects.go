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
    "log"
    "bytes"
    "fmt"
    "time"
)

// Save to Minio Server using PutObject.
func (v *xrayHandlers) saveObjects(data []byte) {   
    // Convert bytes to io.Reader as PutObject expects. Handle error here.
    _, err := v.minioClient.PutObject("alice", fmt.Sprintf("%x/human.jpg", time.Now().UTC().Unix()), bytes.NewReader(data), "image/jpg")
    if err != nil {
        log.Println(err)
        return
    } 
    return
}
