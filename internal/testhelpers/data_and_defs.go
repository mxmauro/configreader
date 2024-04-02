package testhelpers

import (
	"fmt"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------

type TestSettings struct {
	Name              string   `config:"TEST_NAME" validate:"required,min=1,max=64"`
	PtrToName         *string  `config:"TEST_NAME" validate:"required,min=1,max=64"`
	IntegerValue      int      `config:"TEST_INTEGER" validate:"required,number"`
	PtrToIntegerValue *int     `config:"TEST_INTEGER" validate:"required,number"`
	FloatValue        float64  `config:"TEST_FLOAT" validate:"required,number"`
	PtrToFloatValue   *float64 `config:"TEST_FLOAT" validate:"required,number"`

	Server TestSettingsServer

	Node *TestSettingsNode

	MongoDB TestSettingsMongoDB
}

type TestSettingsServer struct {
	Ip               string   `config:"TEST_SERVER_IP" validate:"required,ip"`
	Port             int      `config:"TEST_SERVER_PORT" validate:"required,min=0,max=65535"`
	PoolSize         int      `config:"TEST_SERVER_POOLSIZE" validate:"required,gte=1"`
	AllowedAddresses []string `config:"TEST_SERVER_ALLOWED" is-json:"1" validate:"required"`
}

type TestSettingsNode struct {
	Url      string `config:"TEST_NODE_URL" validate:"required,url"`
	ApiToken string `config:"TEST_NODE_APITOKEN" validate:"required,ascii,min=1,max=64"`
}

type TestSettingsMongoDB struct {
	Url string `config:"TEST_MONGODB_URL" validate:"required,url"`
}

// ------------------------------------------------------------------------------

var GoodSettingsMap = map[string]interface{}{
	"TEST_NAME":            "string test",
	"TEST_INTEGER":         100,
	"TEST_FLOAT":           100.3,
	"TEST_SERVER_IP":       "127.0.0.1",
	"TEST_SERVER_PORT":     "8001",
	"TEST_SERVER_POOLSIZE": 64,
	"TEST_SERVER_ALLOWED":  "[ \"127.0.0.1\", \"::1\" ]",
	"TEST_NODE_URL":        "http://127.0.0.1:8003",
	"TEST_NODE_APITOKEN":   "some-api-access-token",
	"TEST_MONGODB_URL":     "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0",
}

var BadSettingsMap = map[string]interface{}{
	"TEST_NAME":            "string test",
	"TEST_INTEGER":         100,
	"TEST_FLOAT":           "abc",
	"TEST_SERVER_IP":       "127.0.0.1.2",
	"TEST_SERVER_PORT":     8001,
	"TEST_SERVER_POOLSIZE": 64,
	"TEST_SERVER_ALLOWED":  "[ \"127.0.0.1\", \"::1\" ]",
	"TEST_NODE_url":        "http://127.0.0.1:8003",
	"TEST_NODE_apiToken":   "some-api-access-token",
	"TEST_MONGODB_URL":     "mongodb://user:pass@127.0.0.1:27017",
}

// ------------------------------------------------------------------------------

var GoodSettings = TestSettings{
	Name:              "string test",
	PtrToName:         addressOf[string]("string test"),
	IntegerValue:      100,
	PtrToIntegerValue: addressOf[int](100),
	FloatValue:        100.3,
	PtrToFloatValue:   addressOf[float64](100.3),
	Server: TestSettingsServer{
		Ip:               "127.0.0.1",
		Port:             8001,
		PoolSize:         64,
		AllowedAddresses: []string{"127.0.0.1", "::1"},
	},
	Node: &TestSettingsNode{
		Url:      "http://127.0.0.1:8003",
		ApiToken: "some-api-access-token",
	},
	MongoDB: TestSettingsMongoDB{
		Url: "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0",
	},
}

// ------------------------------------------------------------------------------

func ToStringStringMap(srcMap map[string]interface{}) map[string]string {
	dstMap := make(map[string]string)
	for k, v := range srcMap {
		switch _v := v.(type) {
		case string:
			dstMap[k] = _v
		case float32:
			dstMap[k] = fmt.Sprintf("%f", _v)
		case float64:
			dstMap[k] = fmt.Sprintf("%f", _v)
		default:
			dstMap[k] = fmt.Sprintf("%v", _v)
		}
	}
	return dstMap
}

func QuoteValue(v string) string {
	if strings.IndexAny(v, " \t'\"") < 0 {
		return v
	}
	return strconv.Quote(v)
}

func ToKeys(srcMap map[string]interface{}) []string {
	dst := make([]string, 0)
	for k := range srcMap {
		dst = append(dst, k)
	}
	return dst
}

// ------------------------------------------------------------------------------

func addressOf[T any](v T) *T {
	return &v
}
