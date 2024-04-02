package testhelpers

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

// -----------------------------------------------------------------------------

func EnsureAwsEC2Instance(t *testing.T) {
	// Check if the environment variable AWS_EC2_METADATA_DISABLED is set.
	// If it's set to "true", then the instance is not running on EC2.
	if os.Getenv("AWS_EC2_METADATA_DISABLED") != "true" {
		// Make a GET request to the EC2 instance metadata endpoint.
		body, err := queryEc2Metadata("")
		// If the response contains any data, it means the program is running on EC2.
		if err == nil && len(body) > 0 {
			return
		}
	}

	t.Logf("tests are not running in an EC2 instance. skipping....")
	t.SkipNow()
}

func GetEC2InstanceVpcId(t *testing.T) string {
	body, err := queryEc2Metadata("network/interfaces/macs/")
	if err == nil {
		body, _, _ = bytes.Cut(body, []byte{'/'})
		if len(body) > 0 {
			body, err = queryEc2Metadata("network/interfaces/macs/" + string(body) + "/vpc-id")
			if len(body) > 0 {
				return string(body)
			}
		}
	}
	if err != nil {
		t.Fatalf("unable to get AWS EC2 VPC-ID [err=%v]", err)
	}
	t.Fatalf("unable to get AWS EC2 VPC-ID")
	return ""
}

func queryEc2Metadata(path string) ([]byte, error) {
	var body []byte

	resp, err := http.Get("http://169.254.169.254/latest/meta-data/" + path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read the response body which contains metadata information.
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Done
	return body, nil
}
