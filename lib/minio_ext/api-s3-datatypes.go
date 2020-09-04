/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio_ext

import (
	"time"
)

// owner container for bucket owner information.
type owner struct {
	DisplayName string
	ID          string
}

// initiator container for who initiated multipart upload.
type initiator struct {
	ID          string
	DisplayName string
}

// ObjectPart container for particular part of an object.
type ObjectPart struct {
	// Part number identifies the part.
	PartNumber int

	// Date and time the part was uploaded.
	LastModified time.Time

	// Entity tag returned when the part was uploaded, usually md5sum
	// of the part.
	ETag string

	// Size of the uploaded part data.
	Size int64
}

// ListObjectPartsResult container for ListObjectParts response.
type ListObjectPartsResult struct {
	Bucket   string
	Key      string
	UploadID string `xml:"UploadId"`

	Initiator initiator
	Owner     owner

	StorageClass         string
	PartNumberMarker     int
	NextPartNumberMarker int
	MaxParts             int

	// Indicates whether the returned list of parts is truncated.
	IsTruncated bool
	ObjectParts []ObjectPart `xml:"Part"`

	EncodingType string
}