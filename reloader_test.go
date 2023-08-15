package configreader_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/loader"
)

type ReloaderTest struct {
	Value int `json:"value"`
}

//------------------------------------------------------------------------------

func TestReloader(t *testing.T) {
	var value int32 = 1
	var previousValue int = 1

	// Load configuration
	settings, err := configreader.New[ReloaderTest]().
		WithLoader(loader.NewCallback().WithCallback(func(_ context.Context) ([]byte, error) {
			data := fmt.Sprintf(`{ "value": %v }`, atomic.LoadInt32(&value))
			return []byte(data), nil
		})).
		WithReload(time.Second, func(settings *ReloaderTest, loadErr error) {
			if settings.Value == previousValue {
				t.Fatalf("Unexpected value %d (same as previous)", settings.Value)
			}
			previousValue = settings.Value
			t.Logf("Got value %d", settings.Value)
		}).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}
	if settings.Value != 1 {
		t.Fatalf("Unexpected value %d (should be 1)", settings.Value)
	}
	t.Logf("Got value %d", settings.Value)

	for times := 1; times <= 10; times++ {
		<-time.After(time.Second)
		atomic.AddInt32(&value, 1)
	}
}
