package runner

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/altstory/go-log"
)

// Handler 是一个可以执行的函数。
//
// 允许的函数形式包括：
//     - func(ctx context.Context)：      执行一个函数，ctx 由 runner 传入。
//     - func(ctx context.Context) error：与上面的形式类似，只是允许返回一个 error，框架会自动报错。
//
//     - func(ctx context.Context, config *Config)：      这里的 `Config` 是配置文件里面对应的数据结构，
//                                                        Run 会自动解析配置文件并反序列化到 Config 里面去。
//     - func(ctx context.Context, config *Config) error：与上面的形式类似，只是允许返回一个 error，框架会自动报错。
type Handler interface{}

type handler func(ctx context.Context) int
type handlers []handler

var (
	typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
)

// parseHandler 根据反射解析 h 并生成真正的 handler 函数。
//
// section 的值会影响 h 第二个参数在配置中的读取方式。
// 如果 section 为空，那么第二个参数会将整个配置文件内容反序列化到这个参数中。
// 如果 section 不为空，则只会将配置文件中指定 section 反序列化到这个参数中，
// 这会参数读取的更精确。
func parseHandler(section string, h Handler) (handler, error) {
	if h == nil {
		return nil, errors.New("go-runner: handler should not be nil")
	}

	var cfgType reflect.Type
	fn := reflect.ValueOf(h)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		return nil, errors.New("go-runner: handler should be a func")
	}

	inNum := fnType.NumIn()
	outNum := fnType.NumOut()

	if inNum == 0 {
		return nil, errors.New("go-runner: handler must have at least 1 argument")
	} else if inNum > 2 {
		return nil, errors.New("go-runner: handler must have 1 or 2 arguments")
	}

	in0 := fnType.In(0)

	if !in0.Implements(typeOfContext) || !typeOfContext.Implements(in0) {
		return nil, errors.New("go-runner: the type of first argument of handler must be context.Context")
	}

	if inNum == 2 {
		in1 := fnType.In(1)

		if in1.Kind() != reflect.Ptr || in1.Elem().Kind() != reflect.Struct {
			return nil, errors.New("go-runner: the type of second argument in handler must be a pointer to struct")
		}

		cfgType = in1
	}

	if outNum > 1 {
		return nil, errors.New("go-runner: too many values returned by handler")
	}

	if outNum == 1 {
		out := fnType.Out(0)

		if out.Kind() != reflect.Interface {
			return nil, errors.New("go-runner: the return type of handler must be an interface")
		}

		if !out.Implements(typeOfError) || !typeOfError.Implements(out) {
			return nil, errors.New("go-runner: the return type of handler must be error")
		}
	}

	return func(ctx context.Context) int {
		args := make([]reflect.Value, 0, 2)
		args = append(args, reflect.ValueOf(ctx))

		if cfgType != nil {
			// 初始化业务配置。
			arg := reflect.New(cfgType)
			runner := runnerFromContext(ctx)

			if err := runner.Config.Unmarshal(section, arg.Interface()); err != nil {
				log.Errorf(ctx, "err=%v||go-runner: fail to read config", err)
				return ExitCodeInvalidConfig
			}

			args = append(args, arg.Elem())
		}

		returns := fn.Call(args)

		if len(returns) > 0 && returns[0].IsValid() {
			if err, ok := returns[0].Interface().(error); ok && err != nil {
				log.Errorf(ctx, "err=%v||go-runner: fail to call handler", err)
				return ExitCodeHandlerError
			}
		}

		return ExitCodeOK
	}, nil
}

// makeErrorHandler 构建一个专门返回错误的 handler，并输出出错的函数信息。
func makeErrorHandler(skip int, err error) handler {
	caller := findCaller(skip + 1)

	return func(ctx context.Context) int {
		log.Errorf(ctx, "caller=%v||err=%v||go-runner: invalid handler", caller, err)
		return ExitCodeInvalidHandler
	}
}

func findCaller(skip int) string {
	skipStack := skip + 1 // 把当前这个函数也跳过，所以得多 +1。
	caller := "<unknown>"

	if pc, _, _, ok := runtime.Caller(skipStack); ok {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		file = path.Base(file)
		name := f.Name()

		caller = fmt.Sprintf("%v:%v@%v", file, line, name)
	}

	return caller
}

func (h handler) Call(ctx context.Context) (code int) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf(ctx, "go-runner: caught a panic: %v\n%v", r, string(debug.Stack()))
			code = ExitCodeHandlerError
		}
	}()

	code = h(ctx)
	return
}

func (hs handlers) Call(ctx context.Context) int {
	for _, h := range hs {
		if code := h.Call(ctx); code != ExitCodeOK {
			return code
		}
	}

	return ExitCodeOK
}
