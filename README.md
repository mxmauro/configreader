# configreader

Simple configuration reader and validator that accepts a variety of sources.

### Load behavior

The library starts retrieves data from a [source](#sources) and, optionally, expands embedded variables
delimited by `${` and `}`.

The library accepts dotenv files, key/value pairs or JSON objects. Depending on the source type, `C-style` and/or
`# comment` texts will be removed.

## Quick start

1. Import the library

```golang
import (
    "github.com/mxmauro/configreader"
)
```

2. Create a struct that defines your configuration settings and add the required config and validate tags to it.

```golang
type ConfigurationSettings struct {
    Name         string  `config:"TEST_NAME" validate:"required"`
    IntegerValue int     `config:"TEST_INTEGER_VALUE" validate:"required,min=1"`
    FloatValue   float64 `config:"TEST_FLOAT_VALUE"`
}
```

3. Load the application settings like the following example:

```golang
func main() {
    settings, err := configreader.New[ConfigurationSettings]().
        WithLoader(loader.NewFile().WithFilename("./config.env")).
        Load(context.Background())
    if err != nil {
        panic(err.Error())
    }
    ....
}
```

## How to use

Create the configuration reader. Set the name of the struct that defines the configuration definition as the generic
type parameter.

```golang
reader := configreader.New[{structure-name}]()
````

Apply reader options like:

```golang
reader.WithLoader(...).WithMonitor(...)
````
| Method                      | Description                                                                                                             |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------|
| `WithExtendedValidator`     | Sets an optional settings validator callback.                                                                           |
| `WithLoader`                | Sets the content loader. See the [loader section](#loaders) for details.                                                |
| `WithMonitor`               | Sets a monitor that will inform about configuration settings changes. See the [monitor section](#monitors) for details. |
| `WithDisableEnvVarOverride` | Ignore a list of environment variables that can override values.                                                        |

And load the settings:

```golang
ctx := context.Background() // Setup a context
settings, err := reader.Load(ctx)
```

## Loaders

Settings loaders are referenced by importing the following module:

```golang
import (
    "github.com/mxmauro/configreader/loader"
)
```

And then instantiated using the `loader.NewXXX()` functions. 

```golang
ld := loader.NewMemory().WithXXX(...).WithXXX(...)....
reader.WithLoader(ld)
```

You can add more than one loader, overlapping values will be overridden as sources are processed.

See [this document](docs/LOADERS.md) for details about the available loaders.

## Validation

Data validation is executed in two phases. The first inspects `validate` tags using the [Go Playground Validator](https://github.com/go-playground/validator)
library. Please check the full documentation [here](https://pkg.go.dev/github.com/go-playground/validator/v10). 

The second is through the `WithExtendedValidator` method. This library calls the specified function so the developer can
execute further custom checks.

## Monitor

A monitor periodically checks if the configuration setting changed by reloading the settings based on the configuration
reader parameters.

On successful reads, it compares the current values with the previously loaded and, if a different setting is found,
the callback is called with the new values.

If read fails, the callback is called with the error object. 

##### Notes:

* The developer is responsible to notify other application components about the changes.
* If settings are stored in global variables, the developer must ensure synchronized access to them.
* If a reload error occurs, the developer is free to decide the next actions. The monitor will continue trying to
  load settings until explicitly destroyed.

```golang
m := configreader.NewMonitor[{structure-name}](30 * time.Second, func(settings *{structure-name}, loadErr error) {
    // do whatever you need here
})
defer m.Destroy()

settings, err := configreader.New[{structure-name}]().
    WithXXX(...).
    WithMonitor(m).
    Load(...)
```

The callback is called eve

## Variable expansion

Once the key/values are loaded, string values containing expansion macros patterns like `${NAME}` will be automatically
expanded by looking for the specified key.

## Tests

If you want to run the tests of this library, take into consideration the following:

Hashicorp Vault tests requires a running instance launched with the following command:

    vault server -dev -dev-root-token-id="root" -dev-listen-address="0.0.0.0:8200" -dev-no-store-token

If Vault is running on AWS, the EC2 instance requires an IAM role with following policies:

* `ec2:DescribeInstances`
* `iam:GetInstanceProfile` (if IAM Role binding is used)

## LICENSE

See [LICENSE](/LICENSE) file for details.
