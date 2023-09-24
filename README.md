# configreader

Simple configuration reader and validator that accepts a variety of sources.

### Load behavior

The library starts retrieves data from a [source](#sources) and, optionally, expands embedded environment variables
delimited by the `%` character.

Then, `C-style` comments enclosed by the `/* ... */` delimiters or `//` for single-line comments are removed. At this
point it is expected to get a well-formed JSON text which can be validated with a [JSON schema](https://json-schema.org/).

At last, the JSON text is decoded into a structure.


## Quick start

1. Import the library

```golang
import (
    "github.com/mxmauro/configreader"
)
```

2. Create a struct that defines your configuration settings and add the required JSON tags to it.

```golang
type ConfigurationSettings struct {
    Name         string  `json:"name"`
    IntegerValue int     `json:"integerValue"`
    FloatValue   float64 `json:"floatValue"`
}
```

3. Load the application settings like the following example:

```golang
func main() {
    settings, err := configreader.New[ConfigurationSettings]().
        WithLoader(loader.NewFile().WithFilename("./config.json")).
        WithSchema(testhelpers.SchemaJSON).
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
reader.WithLoader(...).WithSchema(...)
````
| Method                  | Description                                                                                                             |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------|
| `WithExtendedValidator` | Sets an optional settings validator callback.                                                                           |
| `WithLoader`            | Sets the content loader. See the [loader section](#loaders) for details.                                                |
| `WithMonitor`           | Sets a monitor that will inform about configuration settings changes. See the [monitor section](#monitors) for details. |
| `WithNoReplaceEnvVars`  | Stops the loader from replacing environment variables that can be found inside.                                         |
| `WithSchema`            | Sets an optional JSON schema validator.                                                                                 |

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

See [this document](docs/LOADERS.md) for details about the available loaders.

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

## Environment variables

Strings passed to a With... method of a loader or found within the loaded content, which contains the `%NAME%` pattern
will try to find the environment variable named `NAME` and replace the tag with its value. Use two (2) consecutive
`%%` characters to replace with one.

## Tests

If you want to run the tests of this library, take into consideration the following:

Hashicorp Vault tests requires a running instance launched with the following command:

    vault server -dev -dev-root-token-id="root" -dev-listen-address="0.0.0.0:8200" -dev-no-store-token

If Vault is running on AWS, the EC2 instance requires an IAM role with following policies:

* `ec2:DescribeInstances`
* `iam:GetInstanceProfile` (if IAM Role binding is used)

## LICENSE

See the [license](LICENSE) file for details.
