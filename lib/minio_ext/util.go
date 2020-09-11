package minio_ext

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v6/pkg/s3utils"
)

// regCred matches credential string in HTTP header
var regCred = regexp.MustCompile("Credential=([A-Z0-9]+)/")

// regCred matches signature string in HTTP header
var regSign = regexp.MustCompile("Signature=([[0-9a-f]+)")

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v interface{}) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// Redact out signature value from authorization string.
func redactSignature(origAuth string) string {
	if !strings.HasPrefix(origAuth, signV4Algorithm) {
		// Set a temporary redacted auth
		return "AWS **REDACTED**:**REDACTED**"
	}

	/// Signature V4 authorization header.

	// Strip out accessKeyID from:
	// Credential=<access-key-id>/<date>/<aws-region>/<aws-service>/aws4_request
	newAuth := regCred.ReplaceAllString(origAuth, "Credential=**REDACTED**/")

	// Strip out 256-bit signature from: Signature=<256-bit signature>
	return regSign.ReplaceAllString(newAuth, "Signature=**REDACTED**")
}

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests.
func closeResponse(resp *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if resp != nil && resp.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

// Verify if input endpoint URL is valid.
func isValidEndpointURL(endpointURL url.URL) error {
	if endpointURL == sentinelURL {
		return ErrInvalidArgument("Endpoint url cannot be empty.")
	}
	if endpointURL.Path != "/" && endpointURL.Path != "" {
		return ErrInvalidArgument("Endpoint url cannot have fully qualified paths.")
	}
	if strings.Contains(endpointURL.Host, ".s3.amazonaws.com") {
		if !s3utils.IsAmazonEndpoint(endpointURL) {
			return ErrInvalidArgument("Amazon S3 endpoint should be 's3.amazonaws.com'.")
		}
	}
	if strings.Contains(endpointURL.Host, ".googleapis.com") {
		if !s3utils.IsGoogleEndpoint(endpointURL) {
			return ErrInvalidArgument("Google Cloud Storage endpoint should be 'storage.googleapis.com'.")
		}
	}
	return nil
}

// getEndpointURL - construct a new endpoint.
func getEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	if strings.Contains(endpoint, ":") {
		host, _, err := net.SplitHostPort(endpoint)
		if err != nil {
			return nil, err
		}
		if !s3utils.IsValidIP(host) && !s3utils.IsValidDomain(host) {
			msg := "Endpoint: " + endpoint + " does not follow ip address or domain name standards."
			return nil, ErrInvalidArgument(msg)
		}
	} else {
		if !s3utils.IsValidIP(endpoint) && !s3utils.IsValidDomain(endpoint) {
			msg := "Endpoint: " + endpoint + " does not follow ip address or domain name standards."
			return nil, ErrInvalidArgument(msg)
		}
	}
	// If secure is false, use 'http' scheme.
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	// Construct a secured endpoint URL.
	endpointURLStr := scheme + "://" + endpoint
	endpointURL, err := url.Parse(endpointURLStr)
	if err != nil {
		return nil, err
	}

	// Validate incoming endpoint URL.
	if err := isValidEndpointURL(*endpointURL); err != nil {
		return nil, err
	}
	return endpointURL, nil
}

var supportedHeaders = []string{
	"content-type",
	"cache-control",
	"content-encoding",
	"content-disposition",
	"content-language",
	"x-amz-website-redirect-location",
	"expires",
	// Add more supported headers here.
}

// isStorageClassHeader returns true if the header is a supported storage class header
func isStorageClassHeader(headerKey string) bool {
	return strings.EqualFold(amzStorageClass, headerKey)
}

// isStandardHeader returns true if header is a supported header and not a custom header
func isStandardHeader(headerKey string) bool {
	key := strings.ToLower(headerKey)
	for _, header := range supportedHeaders {
		if strings.ToLower(header) == key {
			return true
		}
	}
	return false
}

// sseHeaders is list of server side encryption headers
var sseHeaders = []string{
	"x-amz-server-side-encryption",
	"x-amz-server-side-encryption-aws-kms-key-id",
	"x-amz-server-side-encryption-context",
	"x-amz-server-side-encryption-customer-algorithm",
	"x-amz-server-side-encryption-customer-key",
	"x-amz-server-side-encryption-customer-key-MD5",
}

// isSSEHeader returns true if header is a server side encryption header.
func isSSEHeader(headerKey string) bool {
	key := strings.ToLower(headerKey)
	for _, h := range sseHeaders {
		if strings.ToLower(h) == key {
			return true
		}
	}
	return false
}

// isAmzHeader returns true if header is a x-amz-meta-* or x-amz-acl header.
func isAmzHeader(headerKey string) bool {
	key := strings.ToLower(headerKey)

	return strings.HasPrefix(key, "x-amz-meta-") || strings.HasPrefix(key, "x-amz-grant-") || key == "x-amz-acl" || isSSEHeader(headerKey)
}

// sum256 calculate sha256sum for an input byte array, returns hex encoded.
func sum256Hex(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
