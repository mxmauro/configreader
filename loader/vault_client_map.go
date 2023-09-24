package loader

import (
	"crypto/sha256"
	"crypto/tls"
	"runtime"
	"sync"
	"unsafe"
)

// -----------------------------------------------------------------------------

type vaultClientRef uintptr

// -----------------------------------------------------------------------------

var (
	vaultClientMapLock = sync.Mutex{}
	vaultClientMap     = make(map[[32]byte]vaultClientRef)
)

// -----------------------------------------------------------------------------

func (ref vaultClientRef) get() *vaultClient {
	return *(**vaultClient)(unsafe.Pointer(&ref))
}

// -----------------------------------------------------------------------------

func (client *vaultClient) toPtr() vaultClientRef {
	return vaultClientRef(unsafe.Pointer(client))
}

func loadOrStore(client *vaultClient) *vaultClient {
	vaultClientMapLock.Lock()
	defer vaultClientMapLock.Unlock()

	// Check if exists in our weak map
	ref, ok := vaultClientMap[client.hash]
	if ok {
		// Yes
		return ref.get()
	}

	// No, add to it
	vaultClientMap[client.hash] = client.toPtr()

	// And add a finalizer
	runtime.SetFinalizer(client, func(droppingClient *vaultClient) {
		droppingClient.onFinalize()
		unmap(droppingClient)
	})

	// Done
	return client
}

func unmap(client *vaultClient) {
	vaultClientMapLock.Lock()
	defer vaultClientMapLock.Unlock()

	// Avoid race conditions and check we are really freeing the correct instance
	ref, ok := vaultClientMap[client.hash]
	if ok && ref.get() == client {
		delete(vaultClientMap, client.hash)
	}
}

func calculateHash(host string, headers map[string]string, tlsConfig *tls.Config, accessToken string, auth VaultAuthMethod) vaultClientHash {
	const sizeOfUintPtr = unsafe.Sizeof(uintptr(0))
	var hash vaultClientHash

	h := sha256.New()
	_, _ = h.Write([]byte(host))
	if tlsConfig != nil {
		_, _ = h.Write([]byte{'T', 'L', 'S'})
		b := (*[sizeOfUintPtr]byte)(unsafe.Pointer(tlsConfig))[:]
		_, _ = h.Write(b)
	} else {
		_, _ = h.Write([]byte{'N', 'O', 'T', 'L', 'S'})
	}
	if len(headers) > 0 {
		for key, value := range headers {
			_, _ = h.Write([]byte(key))
			_, _ = h.Write([]byte(value))
		}
	}
	if len(accessToken) > 0 {
		_, _ = h.Write([]byte{'T', 'O', 'K', 'E', 'N'})
		_, _ = h.Write([]byte(accessToken))
	} else {
		_, _ = h.Write([]byte{'A', 'U', 'T', 'H'})
		a := auth.hash()
		_, _ = h.Write(a[:])
	}
	copy(hash[:], h.Sum(nil))
	return hash
}
