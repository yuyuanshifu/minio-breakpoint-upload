package minio_ext

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ListObjectParts list all object parts recursively.
func (c Client) ListObjectParts(bucketName, objectName, uploadID string) (partsInfo map[int]ObjectPart, err error) {
	// Part number marker for the next batch of request.
	var nextPartNumberMarker int
	partsInfo = make(map[int]ObjectPart)
	for {
		// Get list of uploaded parts a maximum of 1000 per request.
		listObjPartsResult, err := c.listObjectPartsQuery(bucketName, objectName, uploadID, nextPartNumberMarker, 1000)
		if err != nil {
			return nil, err
		}
		// Append to parts info.
		for _, part := range listObjPartsResult.ObjectParts {
			// Trim off the odd double quotes from ETag in the beginning and end.
			part.ETag = strings.TrimPrefix(part.ETag, "\"")
			part.ETag = strings.TrimSuffix(part.ETag, "\"")
			partsInfo[part.PartNumber] = part
		}
		// Keep part number marker, for the next iteration.
		nextPartNumberMarker = listObjPartsResult.NextPartNumberMarker
		// Listing ends result is not truncated, return right here.
		if !listObjPartsResult.IsTruncated {
			break
		}
	}

	// Return all the parts.
	return partsInfo, nil
}


// listObjectPartsQuery (List Parts query)
//     - lists some or all (up to 1000) parts that have been uploaded
//     for a specific multipart upload
//
// You can use the request parameters as selection criteria to return
// a subset of the uploads in a bucket, request parameters :-
// ---------
// ?part-number-marker - Specifies the part after which listing should
// begin.
// ?max-parts - Maximum parts to be listed per request.
func (c Client) listObjectPartsQuery(bucketName, objectName, uploadID string, partNumberMarker, maxParts int) (ListObjectPartsResult, error) {
	// Get resources properly escaped and lined up before using them in http request.
	urlValues := make(url.Values)
	// Set part number marker.
	urlValues.Set("part-number-marker", fmt.Sprintf("%d", partNumberMarker))
	// Set upload id.
	urlValues.Set("uploadId", uploadID)

	// maxParts should be 1000 or less.
	if maxParts == 0 || maxParts > 1000 {
		maxParts = 1000
	}
	// Set max parts.
	urlValues.Set("max-parts", fmt.Sprintf("%d", maxParts))

	// Execute GET on objectName to get list of parts.
	resp, err := c.executeMethod(context.Background(), "GET", requestMetadata{
		bucketName:       bucketName,
		objectName:       objectName,
		queryValues:      urlValues,
		contentSHA256Hex: emptySHA256Hex,
	})
	defer closeResponse(resp)
	if err != nil {
		return ListObjectPartsResult{}, err
	}
	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			return ListObjectPartsResult{}, httpRespToErrorResponse(resp, bucketName, objectName)
		}
	}
	// Decode list object parts XML.
	listObjectPartsResult := ListObjectPartsResult{}
	err = xmlDecoder(resp.Body, &listObjectPartsResult)
	if err != nil {
		return listObjectPartsResult, err
	}
	return listObjectPartsResult, nil
}
