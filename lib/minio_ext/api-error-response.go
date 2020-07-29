package minio_ext

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	XMLName    xml.Name `xml:"Error" json:"-"`
	Code       string
	Message    string
	BucketName string
	Key        string
	RequestID  string `xml:"RequestId"`
	HostID     string `xml:"HostId"`

	// Region where the bucket is located. This header is returned
	// only in HEAD bucket and ListObjects response.
	Region string

	// Underlying HTTP status code for the returned error
	StatusCode int `xml:"-" json:"-"`
}

// Error - Returns HTTP error string
func (e ErrorResponse) Error() string {
	return e.Message
}
const (
	reportIssue = "Please report this issue at https://github.com/minio/minio/issues."
)
// httpRespToErrorResponse returns a new encoded ErrorResponse
// structure as error.
func httpRespToErrorResponse(resp *http.Response, bucketName, objectName string) error {
	if resp == nil {
		msg := "Response is empty. " + reportIssue
		return ErrInvalidArgument(msg)
	}

	errResp := ErrorResponse{
		StatusCode: resp.StatusCode,
	}

	err := xmlDecoder(resp.Body, &errResp)
	// Xml decoding failed with no body, fall back to HTTP headers.
	if err != nil {
		switch resp.StatusCode {
		case http.StatusNotFound:
			if objectName == "" {
				errResp = ErrorResponse{
					StatusCode: resp.StatusCode,
					Code:       "NoSuchBucket",
					Message:    "The specified bucket does not exist.",
					BucketName: bucketName,
				}
			} else {
				errResp = ErrorResponse{
					StatusCode: resp.StatusCode,
					Code:       "NoSuchKey",
					Message:    "The specified key does not exist.",
					BucketName: bucketName,
					Key:        objectName,
				}
			}
		case http.StatusForbidden:
			errResp = ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       "AccessDenied",
				Message:    "Access Denied.",
				BucketName: bucketName,
				Key:        objectName,
			}
		case http.StatusConflict:
			errResp = ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       "Conflict",
				Message:    "Bucket not empty.",
				BucketName: bucketName,
			}
		case http.StatusPreconditionFailed:
			errResp = ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       "PreconditionFailed",
				Message:    s3ErrorResponseMap["PreconditionFailed"],
				BucketName: bucketName,
				Key:        objectName,
			}
		default:
			errResp = ErrorResponse{
				StatusCode: resp.StatusCode,
				Code:       resp.Status,
				Message:    resp.Status,
				BucketName: bucketName,
			}
		}
	}

	// Save hostID, requestID and region information
	// from headers if not available through error XML.
	if errResp.RequestID == "" {
		errResp.RequestID = resp.Header.Get("x-amz-request-id")
	}
	if errResp.HostID == "" {
		errResp.HostID = resp.Header.Get("x-amz-id-2")
	}
	if errResp.Region == "" {
		errResp.Region = resp.Header.Get("x-amz-bucket-region")
	}
	if errResp.Code == "InvalidRegion" && errResp.Region != "" {
		errResp.Message = fmt.Sprintf("Region does not match, expecting region ‘%s’.", errResp.Region)
	}

	return errResp
}

func ToErrorResponse(err error) ErrorResponse {
	switch err := err.(type) {
	case ErrorResponse:
		return err
	default:
		return ErrorResponse{}
	}
}
// ErrInvalidArgument - Invalid argument response.
func ErrInvalidArgument(message string) error {
	return ErrorResponse{
		Code:      "InvalidArgument",
		Message:   message,
		RequestID: "minio",
	}
}

// ErrEntityTooLarge - Input size is larger than supported maximum.
func ErrEntityTooLarge(totalSize, maxObjectSize int64, bucketName, objectName string) error {
	msg := fmt.Sprintf("Your proposed upload size ‘%d’ exceeds the maximum allowed object size ‘%d’ for single PUT operation.", totalSize, maxObjectSize)
	return ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "EntityTooLarge",
		Message:    msg,
		BucketName: bucketName,
		Key:        objectName,
	}
}

// ErrEntityTooSmall - Input size is smaller than supported minimum.
func ErrEntityTooSmall(totalSize int64, bucketName, objectName string) error {
	msg := fmt.Sprintf("Your proposed upload size ‘%d’ is below the minimum allowed object size ‘0B’ for single PUT operation.", totalSize)
	return ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "EntityTooSmall",
		Message:    msg,
		BucketName: bucketName,
		Key:        objectName,
	}
}
