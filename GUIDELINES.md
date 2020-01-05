# Project Guidelines
Miscellaneous conventions and guidelines for writing Enseada code

<!-- toc -->

- [Project Structure](#project-structure)
  - [Directory Structure](#directory-structure)
  - [Go Packages](#go-packages)
- [Code organization](#code-organization)
  - [Package Per Context](#package-per-context)
  - [No Globals](#no-globals)
  - [Interfaces](#interfaces)
- [Dependency Injection](#dependency-injection)
- [Configuration](#configuration)
- [RPC](#rpc)

<!-- tocstop -->

## Project Structure

### Directory Structure

Enseada follows the classic Golang layout:

```bash
.
├── cmd
│   ├── enseada         # client CLI
│   └── enseada-server  # server CLI
├── conf                # miscellaneous configurtion files
├── docs                # markdowns and generated documentation files
├── examples            # example files (e.g. local certificates)
├── internal            # source code not to be used by third parties    
├── pkg                 # source code that can be imported by third parties
├── rpc                 # protobuf definitions and generated code
└── web                 # frontend related stuff (JavaScript, CSS, templates, etc...)
``` 

The division between `cmd`, `pkg`, and `internal` is really important.   

Most of the code resides in `internal` and, as such, can be modified and broken without prior warning. 
Code in this folder should **NEVER** import packages from `pkg` or `cmd`, only third party dependencies.  
_EXCEPTION_: the `log` package. This is just an interface and it needs to be accessed pretty much everywhere. It will eventually be extracted
in a library.

Code residing in `pkg` is safe for external usage. Its API should be stable (following the project
SemVer) and is the only code allowed to import packages from `internal`. This is also where [Dependency Injection](#dependency-injection)
constructor functions are defined.

`cmd` executables are treated as third party consumers, meaning they can only import code from `pkg`, **NOT** from `internal`. This helps
in dogfooding the implementation so that the interface can stay stable.

If in doubt, put new code in `internal`. It is easier to promote code to `pkg` than to break
existing consumers by changing stable interfaces.

### Go Packages

We limit package depth to 1. This means there can be at most one level of nesting inside `cmd` and `internal`

```bash
internal/auth/oauth_client.go GOOD

intenal/auth/oauth/client.go  BAD
```

Top level commands in `cmd` executables should have ideally only one package, `main`. Nested packages are allowed for subcommands, but ideally should never
go beyond depth 1.

```bash
cmd/enseada/version.go            GOOD

cmd/enseada/get/repository.go     OK

cmd/enseada/get/repository/run.go BAD
```

## Code organization

### Package Per Context

Code is grouped in packages that share a common context based on the [Domain Driven Design definition](https://martinfowler.com/bliki/BoundedContext.html).

This means that there is no `models` or `services` package, but instead all the code related to a context lives in the same package.
So for example, all the users related code can live in the `users` package:
```bash
internal
└── users
    ├── model.go
    ├── model_test.go
    ├── password.go
    ├── password_test.go
    ├── service.go
    └── service_test.go
```

The only exception are RPC structs (for example, a `User` protobuf message). The generated code lives in the `rpc` folder so that clients
can cleanly import it without binding to the internal implementation.

### No Globals

This should come to no surprise, no code inside `pkg` or `internal` can define or use globals. This means no shared instances, no shared variables and no `init()` funcs.
Only exceptions to this rules are `consts` and `vars` that define common values, for example
error messages, `kind` values and hardcoded names (e.g. database names).

`cmd` executables are allowed to use `init()` functions and global instances, albeit it is discouraged. A common example is defining Cobra commands as global vars.

Instead, the code should be modularized in side-effect-free functions, and state should be encapsulated in structs. See [Dependency Injection](#dependency-injection) for more on that.  

### Interfaces

    Accept interfaces, return structs
A popular practice in Golang is to never accept structs as function parameters (although there are cases where this is allowed, for example for Options structs) and to instead accept interfaces.
This allows to decouple behaviours from their implementation.

This rule leads to another common practice: defining interface at point of use. For example, if we have a `Greeter` interface and a `InMemoryGreeter` implementation struct,
the interface should not be defined alongside the struct, but where the interface is actually used.

```go
package greeterimpl

import "fmt"

type InMemoryGreeter struct {}

func (g *InMemoryGreeter) Greet(name string) string {
    return fmt.Sprintf("Hello %s!", name)
}
```

```go
package greeter

import "fmt"

type Greeter interface {
    Greet(name string) string
}

func Hello(g Greeter) {
    fmt.Println(g.Greet("Enseada"))
}
```

This removes the dependency between package `greeter` and `greeterimpl`, simplifying code and reducing risk of breakage. See [this article](https://blog.chewxy.com/2018/03/18/golang-interfaces/) for a more in-depth explanation.

## Dependency Injection

Since globals are not allowed, DI is used instead to pass dependencies to structs.
Enseada uses a custom library for application lifecycle management (located in [pkg/app](./pkg/app)) that encapsulates code into modules.

The minimum unit of management is a `Module`. A module is defined by the interface `app.Module` and represents
a unit of code that can be initialized, started and stopped as a whole. Each module can export different components that can be injected into other modules. For
example, the [http module](./pkg/http/module.go) exports an `*echo.Echo` instance to be used by the other modules.

All modules are then combined into an `app.App`. An app is a structure that will start and stop modules as a cohesive group.
The app will also handle errors and panics as well as emit application events.

A module can also execute code in response to application events (like `BeforeApplicationStart` or `AfterApplicationStop`) by implementing
`app.LifecycleAware`. Registered listeners are executed in parallel, each one in its own goroutine, so be aware that if initialization must happen
in a specific order (because of module dependencies), this initialization must happen in the module constructor. For this reason it is strongly recommended
accepting a `context.Context` as the constructor first param (see for example the [auth module](./pkg/auth/module.go)). 

Each module is started in its own goroutine. This means that:

1. A module `Start()` method CAN be blocking. If `Start()` is NOT blocking, returning `nil` indicates a successful start. Returning an error
will cause the entire app to halt and all modules to be stopped.
2. While modules are started in the order they are registered, there is NO guarantee that a module
will actually start before another, since they are executed in parallel. If you need to run code before or after another dependency module either
put it in the module constructor.

**WARNING** 

DI should **ONLY** be used in `cmd` executables. No injection configuration should be present
in `pkg` or `internal` code. Wiring up components is the consumer responsibility, not the core application.

## Configuration

Like dependency injection, configuration management is the consumer responsibility. Core application code should accepts params as function arguments or options structs and
never read the application environment directly.

Enseada `cmd` executables use [Viper](https://github.com/spf13/viper/) to manage configuration from different sources (flags, files, environment variabiles) and 
pass it to the actual code. 

## RPC

Enseada exposes a [Twirp](https://twitchtv.github.io/twirp/docs/intro.html) API and as such generates a fair bunch of code from [Protocol Buffers](https://developers.google.com/protocol-buffers/).
The message definitions can be found in the `rpc` folder, divided by context and version (e.g. `rpc/auth/v1`) following [Uber's V2 Style Guide](https://github.com/uber/prototool/blob/dev/style/README.md).
Linting, compiling and checking for breaking changes is done using [prototool](https://github.com/uber/prototool/). Generated Go code is placed alongside the relevant `.proto` files so that clients can avoid importing core code.

```go

package myclient

import (

"context"
mavenv1beta1 "github.com/enseadaio/enseada/rpc/maven/v1beta1"
"net/http"
)

func Run(ctx context.Context) error {
    client := mavenv1beta1.NewMavenAPIProtobufClient("http://localhost:9623", &http.Client{})
    res, err := client.ListRepos(ctx, &mavenv1beta1.ListReposRequest{})
    if err != nil {
        return err
    }
    // do stuff with response
    return nil
}
```