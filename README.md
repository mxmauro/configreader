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
| Method                  | Description                                                                           |
|-------------------------|---------------------------------------------------------------------------------------|
| `WithExtendedValidator` | Sets an optional settings validator callback.                                         |
| `WithLoader`            | Sets the content loader. See the [loader section](#loaders) for details.              |
| `WithReload`            | Sets a polling interval and a callback to call if the configuration settings changes. |
| `WithNoReplaceEnvVars`  | Stops the loader from replacing environment  variables that can be found inside       |
| `WithSchema`            | Sets an optional JSON schema validator.                                               |

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

## Environment variables

Strings passed to a With... method of a loader or found within the loaded content, which contains the `%NAME%` pattern
will try to find the environment variable named `NAME` and replace the tag with its value. Use two (2) consecutive
`%%` characters to replace with one.

## LICENSE

See the [license](LICENSE) file for details.
