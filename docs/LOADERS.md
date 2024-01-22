## Configuration loaders

The following loaders are available as sources of configuration settings.

### Memory

```golang
loader.NewMemory()
loader.NewMemoryFromEnvironmentVariable(Name string)
```

Creates a loader from a value stored in memory. Available loader options:

| Method     | Description                                         |
|------------|-----------------------------------------------------|
| `WithData` | Sets the data to return when the content is loaded. |


### Callback

```golang
loader.NewCallback()
```

Creates a loader which calls a callback function which, in turn, returns some content. Available loader options:

| Method         | Description                 |
|----------------|-----------------------------|
| `WithCallback` | Sets the callback function. |

### File

```golang
loader.NewFile()
loader.NewFileFromCommandLine(CmdLineParameter *string, CmdLineParameterShort *string)
loader.NewFileFromEnvironmentVariable(Name string)
```

Creates a loader that reads data from a file. Available loader options:

| Method         | Description        |
|----------------|--------------------|
| `WithFilename` | Sets the filename. |

### Http

```golang
loader.NewHttp()
```

Creates a loader that reads data from a website. Available loader options:

| Method            | Description                                      |
|-------------------|--------------------------------------------------|
| `WithURL`         | Sets the options from the provided url.          |
| `WithHost`        | Sets the host address and, optionally, the port. |
| `WithPath`        | Sets the URL path.                               |
| `WithQuery`       | Sets the query parameters.                       |
| `WithQueryItem`   | Sets a single query parameter.                   |
| `WithCredentials` | Sets the username and password.                  |
| `WithDefaultTLS`  | Sets a default `tls.Config` object.              |
| `WithTLS`         | Sets a `tls.Config` object.                      |
| `WithHeaders`     | Sets the request headers.                        |
| `WithHeaderItem`  | Sets a single request header.                    |

### Vault

```golang
loader.NewVault()
```

Creates a loader that reads data from [Hashicorp Vault](https://www.vaultproject.io/). Available loader options:

| Method            | Description                                      |
|-------------------|--------------------------------------------------|
| `WithURL`         | Sets the options from the provided url.          |
| `WithAuth`        | Sets the authorization method to use.            |
| `WithAccessToken` | Sets the access token to use as authorization.   |
| `WithHost`        | Sets the host address and, optionally, the port. |
| `WithPath`        | Sets the URL path.                               |
| `WithDefaultTLS`  | Sets a default `tls.Config` object.              |
| `WithTLS`         | Sets a `tls.Config` object.                      |
| `WithHeaders`     | Sets the request headers.                        |
| `WithHeaderItem`  | Sets a single request header.                    |

[This document](VAULT.md) describes available authorization methods for use in the Vault loader.

### Auto-detect

```golang
loader.NewAutoDetect(opts...)
```

Creates a loader that tries to auto-detect the source of the data to read.

Depending on the provided options, it tries to check if the origin is passed with a command-line parameter or an
environment variable.

Then, a Hashicorp Vault, an HTTP or a File loader is created and returned.
