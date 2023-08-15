package testhelpers

//------------------------------------------------------------------------------

type TestSettings struct {
	Name         string  `json:"name"`
	IntegerValue int     `json:"integerValue"`
	FloatValue   float64 `json:"floatValue"`

	Server TestSettingsServer `json:"server"`

	Node TestSettingsNode `json:"node"`

	MongoDB TestSettingsMongoDB `json:"mongodb"`
}

type TestSettingsServer struct {
	Ip               string   `json:"ip"`
	Port             int      `json:"port"`
	PoolSize         int      `json:"poolSize"`
	AllowedAddresses []string `json:"allowedAddresses"`
}

type TestSettingsNode struct {
	Url      string `json:"url"`
	ApiToken string `json:"apiToken"`
}

type TestSettingsMongoDB struct {
	Url string `json:"url"`
}

//------------------------------------------------------------------------------

var SchemaJSON = `{
	"$schema": "http://json-schema.org/draft-07/schema",
	"$id": "http://example.com/example.json",
	"title": "Sample.",
	"description": "This is a sample configuration settings.",
	"type": "object",
	"required": [
		"name",
		"integerValue",
		"floatValue",
		"server",
		"node",
		"mongodb"
	],
	"properties": {
		"name": {
			"$id": "#/properties/name",
			"title": "name",
			"description": "A random string value.",
			"type": "string",
			"minLength": 1
		},
		"integerValue": {
			"$id": "#/properties/integerValue",
			"title": "integerValue",
			"description": "A random integer value.",
			"type": "integer"
		},
		"floatValue": {
			"$id": "#/properties/floatValue",
			"title": "floatValue",
			"description": "A random float value.",
			"type": "number"
		},
		"server": {
			"$id": "#/properties/server",
			"title": "Server information",
			"description": "Server parameters the application will mount.",
			"type": "object",
			"required": [
				"ip",
				"port"
			],
			"properties": {
				"ip": {
					"$id": "#/properties/server/properties/ip",
					"title": "Bind address",
					"description": "The bind address to listen for incoming connections.",
					"type": "string",
					"anyOf": [
						{ "format": "ipv4" },
						{ "format": "ipv6" }
					],
					"default": "127.0.0.1"
				},
				"port": {
					"$id": "#/properties/server/properties/port",
					"title": "Listen port",
					"description": "The port to use to listen for incoming connections.",
					"type": "integer",
					"minimum": 0,
					"maximum": 65535
				}
			}
		},

		"node": {
			"$id": "#/properties/node",
			"title": "Algorand's Node settings",
			"description": "Indicates the connection settings to use to connect to Algorand's Node.",
			"type": "object",
			"required": [
				"url",
				"apiToken"
			],
			"properties": {
				"url": {
					"$id": "#/properties/node/properties/url",
					"title": "Url",
					"description": "Specifies the node URL.",
					"type": "string",
					"pattern": "^https?:\\/\\/([^:/?#]+)(\\:\\d+)$"
				},
				"apiToken": {
					"$id": "#/properties/node/properties/apiToken",
					"title": "API Access Token",
					"description": "Specifies the access token to use.",
					"type": "string",
					"minLength": 1
				}
			}
		},

		"mongodb": {
			"$id": "#/properties/mongodb",
			"title": "MongoDB database settings",
			"description": "Indicates the database connection settings to use.",
			"type": "object",
			"required": [
				"url"
			],
			"properties": {
				"url": {
					"$id": "#/properties/mongodb/properties/url",
					"title": "MongoDB connection URL",
					"description": "Indicates the connection settings to use to connect to MongoDB.",
					"type": "string",
					"pattern": "^mongodb:\\/\\/(([^:@]*\\:[^@]*)@)?([^:/?#]+)(\\:\\d+)?(\/[a-zA-Z0-9_]*)\\?replSet\\=\\w+",
					"examples": [
						"mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0"
					]
				}
			}
		}
	},

	"additionalProperties": true
}`

var GoodSettingsJSON = `{
	"name": "string test",
	"integerValue": 100,
	"floatValue": 100.3,

	"server": {
		"ip": "127.0.0.1",
		"port": 8001,
		"poolSize": 64,
		"allowedAddresses": [
			"127.0.0.1", "::1"
		]
	},

	"node": {
		"url": "http://127.0.0.1:8003",
		"apiToken": "some-api-access-token"
	},

	"mongodb": {
		"url": "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0"
	}
}`

var BadSettingsJSON = `{
	"name": 1,
	"integerValue": "100",
	"floatValue": "100.3",

	"server": {
		"ip": "my_localhost",
		"port": 8001,
		"poolSize": 64
	},

	"node": {
		"url": "http://127.0.0.1:8003",
		"apiToken": "some-api-access-token"
	},

	"mongodb": {
		"url": "mongodb://user:pass@127.0.0.1:27017"
	}
}`

var GoodSettings = TestSettings{
	Name:         "string test",
	IntegerValue: 100,
	FloatValue:   100.3,
	Server: TestSettingsServer{
		Ip:               "127.0.0.1",
		Port:             8001,
		PoolSize:         64,
		AllowedAddresses: []string{"127.0.0.1", "::1"},
	},
	Node: TestSettingsNode{
		Url:      "http://127.0.0.1:8003",
		ApiToken: "some-api-access-token",
	},
	MongoDB: TestSettingsMongoDB{
		Url: "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0",
	},
}
