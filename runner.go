package runner

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"

	"github.com/altstory/go-config"
	"github.com/altstory/go-log"
)

const (
	// ExitCodeOK 是正常退出时候的错误码。
	ExitCodeOK = iota

	// ExitCodeInvalidHandler 是系统错误码，表示执行的 handler 类型非法。
	ExitCodeInvalidHandler

	// ExitCodeInvalidConfig 是配置文件解析失败时候的错误码。
	ExitCodeInvalidConfig

	// ExitCodeHandlerError 是 handler 执行出错或者 panic 返回的错误码。
	ExitCodeHandlerError
)

var (
	flagConfig  = flag.String("config", "./conf/service.conf", "Set config file for this server.")
	flagVersion = flag.Bool("version", false, "Display version of this server.")
)

type keyRunnerContextType struct{}

var (
	keyRunnerContext keyRunnerContextType
)

type runnerContext struct {
	Config *config.Config
}

// Main 是整个框架的启动入口，这个函数永远不会返回。
func Main() {
	flag.Parse()

	// 读取环境信息。
	parseMetaInfo()

	if *flagVersion {
		meta := Meta()
		fmt.Printf("%s rev:%s\n", meta.Project, meta.GitRevision)
		return
	}

	os.Exit(run())
	panic("never reach here")
}

const envRunnerExtConfig = "ALTSTORY_RUNNER_EXT_CONFIG"

func run() (code int) {
	runner := &runnerContext{}
	ctx := context.WithValue(context.Background(), keyRunnerContext, runner)

	// 先自动初始化日志。
	path := *flagConfig
	c, err := config.LoadFile(path)

	if err != nil {
		log.Errorf(ctx, "err=%v||config=%v||go-runner: fail to parse config file", err, path)
		return ExitCodeInvalidConfig
	}

	// 如果有环境变量设置了额外追加的配置文件，加载这个配置文件。
	extPath, exists := os.LookupEnv(envRunnerExtConfig)

	if exists && extPath != "" {
		err = c.LoadExt(extPath)

		if err != nil {
			log.Errorf(ctx, "err=%v||config=%v||ext_config=%v||go-runner: fail to parse extension config file", err, path, extPath)
			return ExitCodeInvalidConfig
		}
	}

	runner.Config = c
	var logConfig log.Config

	if err := runner.Config.Unmarshal("log", &logConfig); err != nil {
		log.Errorf(ctx, "err=%v||go-runner: fail to read config", err)
		return ExitCodeInvalidConfig
	}

	// 初始化日志。
	updatePackagePrefix(&logConfig)
	log.Init(&logConfig)
	defer func() {
		log.Warnf(ctx, "code=%v||go-runner: server is exiting", code)
		log.Flush()
	}()

	// 配置日志切分。
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)
	go func() {
		for range sig {
			log.Rotate()
		}
	}()
	defer func() {
		signal.Stop(sig)
		close(sig)
	}()

	defer runExitHandlers(ctx)

	if code = runConfigHandlers(ctx); code != ExitCodeOK {
		return
	}

	if code = runClients(ctx); code != ExitCodeOK {
		return
	}

	if code = runStartHandlers(ctx); code != ExitCodeOK {
		return
	}

	code = runServers(ctx)
	return
}

// updatePackagePrefix 在调用栈中找到第一个跟当前 PkgPath 不一样的包路径作为包前缀信息，
// 这样可以简化日志里面的调用栈信息，在不损失信息量的前提下缩减日志文件体积。
func updatePackagePrefix(c *log.Config) {
	if c.PackagePrefix != "" {
		return
	}

	currentPkg := reflect.TypeOf((*Handler)(nil)).Elem().PkgPath()

	pcs := make([]uintptr, 10)
	l := runtime.Callers(3, pcs)
	pcs = pcs[:l]

	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)

		if f == nil {
			continue
		}

		name := f.Name()
		idx := strings.LastIndex(name, "/")

		if idx < 0 {
			continue
		}

		dotIdx := strings.Index(name[idx:], ".")

		if dotIdx < 0 {
			continue
		}

		prefix := name[:idx+dotIdx]

		if prefix != currentPkg {
			c.PackagePrefix = prefix
			return
		}
	}
}

func runnerFromContext(ctx context.Context) *runnerContext {
	return ctx.Value(keyRunnerContext).(*runnerContext)
}
