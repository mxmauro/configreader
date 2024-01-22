//go:build !windows

package loader

// -----------------------------------------------------------------------------

var k8sTokenFiles = []string{
	"/var/run/secrets/kubernetes.io/serviceaccount/token",
	"/run/secrets/kubernetes.io/serviceaccount/token",
}
