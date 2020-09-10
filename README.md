# go-runner：Go 框架总入口 #

`go-runner` 设计初衷是提供一个 all-in-one 入口，尽可能减少使用者的思考负担。

由于 v1 版本相对之前版本有非常大的变化，详细迁移方法请参考 [`v0` 到 `v1` 迁移指南](migrate-from-v0-to-v1.md)。

## 使用方法 ##

一般来说，业务的 `main` 函数里面只需要注册各种启动函数，最终调用 `Main` 函数即可完成启动。

所有 altstory 中的 client/server 库都已经完成跟 `go-runner` 的集成，一般来说无需操心初始化工作即可直接使用。
只要业务代码中使用了这些库，则这些库就会在服务启动时候自动初始化，无需显式的写任何初始化代码。

```go
import (
    "github.com/altstory/go-http/server"
    "github.com/altstory/go-runner"
    "github.com/project/server/routes"
)

func main() {
    // 将所有业务 route 注册到 http server 里面，
    // 一旦注册了 routes，http server 就被自动启动起来了。
    server.AddRoutes(routes.Routes)

    // 启动服务。
    runner.Main()
}
```

### 自定义业务配置 ###

如果业务代码中需要使用一些从配置文件中读出来的信息，推荐用以下模式来实现。

首先，在项目目录中创建一个 `./config/` 目录。

假设这个配置项名字叫做 `resource`，那么创建文件 `./config/resource.go` 并写下如下代码。

```go
package config

import "github.com/altstory/go-runner"

// ResourceType 是资源相关接口的配置。
type ResourceType struct {
    Foo string `config:"foo"`
    Bar int `config:"bar"`
}

// Resource 是资源相关配置的值。
var Resource *ResourceType

func init() {
    // 注册自动解析配置的函数，在服务启动后即可自动从配置文件中读取对应内容。
    runner.LoadConfig("resource", &Resource)
}
```

### 命令行参数 ###

默认情况下，通过 `Main` 启动的服务会提供以下参数：

* `-config`：指定配置文件，默认是 `./conf/service.conf`；
* `-version`：返回当前服务版本信息，这需要 CI 系统配合生成 `.meta.json`。

### 修改日志配置 ###

`go-runner` 会假定配置文件中 `[log]` 的部分是日志配置。

```ini
# [log] 部分对应 log.Config 的各个配置项，可以用来控制日志的细节配置。
[log]
log_level = "debug"
```

### 通过环境变量追加配置 ###

为了方便管理生成出来的 docker 镜像里面的配置，框架提供了一个环境变量 `ALTSTORY_RUNNER_EXT_CONFIG` 来设置一个额外的配置文件，用于覆盖镜像里自带的配置。

例如，有一个额外的配置文件 `path/to/service-ext.conf`。

```ini
# 可以删除指定的配置项。
_deletes = ['http.server.debug', 'other.key']

# 也可以覆盖指定的配置项。
[log]
log_level = "info"
```

在启动服务前，设置以下环境变量即可将这个配置追加到默认配置里面去。

    docker run --env ALTSTORY_RUNNER_EXT_CONFIG=path/to/service-ext.conf

### 获取环境信息 ###

根据公司的 CI 脚本设计，我们会在每个通过 CI build 的 docker 镜像里面放入一个 `.meta.json` 文件，用来告诉服务当前环境信息。如果服务希望读取这个文件里面的信息，可以通过调用 `Meta` 方法来获得所有数据。

```go
// 任何业务代码里面都可以直接调用这个函数。
meta := runner.Meta()
fmt.Println(meta.Project)
```

### 注册启动和退出函数 ###

业务代码可以通过 `OnStart` 和 `OnExit` 来注册启动和退出函数，它们的调用时机是：

* `OnStart` 会在所有 client 初始化完成、所有 server 还未初始化的时候调用，注册到 `OnStart` 的函数的执行顺序是不保证有序的，业务不应该依赖这个执行顺序来执行业务代码。
* `OnExit` 会在服务退出时候调用，注册到 `OnExit` 的函数的执行顺序是不保证有序的，业务不应该依赖这个执行顺序来执行业务代码。

这两个函数都应该在 `init` 中调用。

```go
func init() {
    runner.OnStart(func(ctx context.Context) error {
        // 业务初始化代码……

        // nil 表示初始化正常，继续执行其他的启动过程。
        // 如果返回错误，则会阻止服务继续启动，程序会退出和报错。
        return nil
    })

    runner.OnExit(func(ctx context.Context) {
        // 业务退出代码……
    })
}
```
