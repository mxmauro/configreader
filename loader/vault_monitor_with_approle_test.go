package loader_test

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

// -----------------------------------------------------------------------------

type MultiMonitorTest struct {
	Value int `config:"TEST_VALUE"`
}

// -----------------------------------------------------------------------------

func TestVaultLoaderMonitoringWithAppRoleAuth(t *testing.T) {
	// Check if vault is running
	vaultAddr := testhelpers.EnsureVaultAvailability(t)

	// Create vault accessor with root privileges
	client := testhelpers.CreateRootVaultClient(t, vaultAddr)

	// Add AppRole auth engine
	testhelpers.EnableAppRoleAuthEngine(t, client)

	// Create the role, get its ID and create a secret id
	testhelpers.CreateAppRoleRoleForReadSecretPolicy(t, client)
	roleId := testhelpers.GetAppRoleRoleID(t, client)
	secretId := testhelpers.CreateAppRoleSecretID(t, client)

	// Write the secrets
	testhelpers.WriteVaultSecrets(t, client, "settings-1", createMultiMonitorSettingsWithValue(1000))
	testhelpers.WriteVaultSecrets(t, client, "settings-2", createMultiMonitorSettingsWithValue(2000))

	// cT := testhelpers.NewConcurrentT(t)

	// This test will last some time
	ctx, cancelCtx := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelCtx()

	ae := testhelpers.NewAtomicError()
	wg := sync.WaitGroup{}
	wg.Add(2)
	for workerNo := 1; workerNo <= 2; workerNo += 1 {
		go func(workerNo int) {
			defer wg.Done()

			localCtx, cancelLocalCtx := context.WithCancel(context.Background())
			defer cancelLocalCtx()

			var previousValue = workerNo * 1000

			settingsMonitor := configreader.NewMonitor[MultiMonitorTest](time.Second, func(settings *MultiMonitorTest, loadErr error) {
				if loadErr != nil {
					ae.Set(loadErr)
					cancelLocalCtx()
					return
				}
				if settings.Value == previousValue {
					ae.Set(fmt.Errorf("unexpected value %d (same as previous)", settings.Value))
					cancelLocalCtx()
					return
				}
				previousValue = settings.Value
				t.Logf("Worker #%v got value %d", workerNo, settings.Value)
			})
			defer settingsMonitor.Destroy()

			// Load configuration from vault using app-role
			_, err := configreader.New[MultiMonitorTest]().
				WithLoader(loader.NewVault().
					WithHost(vaultAddr).
					WithPath(testhelpers.PathFromSecretKey("settings-" + strconv.Itoa(workerNo))).
					WithAuth(loader.NewVaultAppRoleAuthMethod().
						WithRoleId(roleId).
						WithSecretId(secretId))).
				WithMonitor(settingsMonitor).
				Load(context.Background())
			if err != nil {
				ae.Set(err)
				return
			}

			for check := 1; check <= 10; check += 1 {
				select {
				case <-ctx.Done():
					return
				case <-localCtx.Done():
					return
				case <-time.After(time.Duration(200+(workerNo*400)) * time.Millisecond):
					testhelpers.WriteVaultSecrets(t, client, "settings-"+strconv.Itoa(workerNo),
						createMultiMonitorSettingsWithValue((workerNo*1000)+check))
				}
			}
		}(workerNo)
	}

	wg.Wait()

	if ae.Err() != nil {
		t.Fatalf("%v", ae.Err())
	}
}

// -----------------------------------------------------------------------------

func createMultiMonitorSettingsWithValue(newValue int) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["TEST_VALUE"] = newValue
	return ret
}
