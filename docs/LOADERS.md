## Configuration loaders

The following loaders are available as sources of configuration settings.

### Memory

Creates a loader from a value stored in memory.

```golang
loader.NewMemory()
loader.NewMemoryFromEnvironmentVariable(Name string)
```

Available loader options:

| Method     | Description                                         |
|------------|-----------------------------------------------------|
| `WithData` | Sets the data to return when the content is loaded. |

### Callback

Creates a loader which calls a callback function which, in turn, returns some content.

```golang
loader.NewCallback()
```

Available loader options:

| Method         | Description                 |
|----------------|-----------------------------|
| `WithCallback` | Sets the callback function. |

### File

Creates a loader that reads data from a file. Data can be in [DotEnv](https://www.dotenv.org/docs/security/env) format
or [JSON](https://www.json.org/).

```golang
loader.NewFile()
loader.NewFileFromCommandLine(CmdLineParameter *string, CmdLineParameterShort *string)
loader.NewFileFromEnvironmentVariable(Name string)
```

Available loader options:

| Method         | Description        |
|----------------|--------------------|
| `WithFilename` | Sets the filename. |

### Http

Creates a loader that reads data from a website. Like the file loader, data can be in DotEnv or JSON format.

```golang
loader.NewHttp()
```

Available loader options:

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

Creates a loader that reads data from [Hashicorp Vault](https://www.vaultproject.io/).

```golang
loader.NewVault()
```

Available loader options:

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
