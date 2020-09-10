package runner

import (
	"context"

	"github.com/altstory/go-config"
	"github.com/altstory/go-log"
)

var configHandlers handlers

// LoadConfig 将 v 注册到启动逻辑里面，一旦配置文件读取之后，会从指定的 secion 给 v 赋值。
//
// 例如：
//
//     type FooConfig struct {
//         Bar int `config:"bar"`
//     }
//
//     // 声明一个全局变量，用来方便业务代码读取配置。
//     var Foo *FooConfig
//
//     func init() {
//         // 在自启动函数里面注册 Foo，这样服务启动后就会自动从配置里面读取 Foo 的值。
//         // 假如读取过程中出现任何问题，比如配置里面没有 foo 这个字段，Foo 为 nil。
//         runner.LoadConfig("foo", &Foo)
//     }
func LoadConfig(section string, v interface{}) {
	LoadConfigFile("", section, v)
}

// LoadConfigFile 将 v 注册到启动逻辑里面，一旦配置文件读取之后，会从指定的 secion 给 v 赋值。
// 跟 LoadConfig 不一样的是，通过指定 path，可以指定一个跟默认配置文件不一样的配置文件。
func LoadConfigFile(path string, section string, v interface{}) {
	configHandlers = append(configHandlers, func(ctx context.Context) int {
		runner := runnerFromContext(ctx)
		c := runner.Config

		if path != "" {
			conf, err := config.LoadFile(path)

			if err != nil {
				log.Errorf(ctx, "err=%v||config=%v||go-runner: fail to parse config file", err, path)
				return ExitCodeInvalidConfig
			}

			c = conf
		}

		err := c.Unmarshal(section, v)

		if err != nil {
			log.Errorf(ctx, "err=%v||go-runner: fail to read config", err)
			return ExitCodeInvalidConfig
		}

		return ExitCodeOK
	})
}

func runConfigHandlers(ctx context.Context) int {
	return configHandlers.Call(ctx)
}
