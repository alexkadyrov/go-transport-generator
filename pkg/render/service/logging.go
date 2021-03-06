package service

import (
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/wildberries-ru/go-transport-generator/pkg/api"
)

const loggingTpl = `// Package {{.PkgName}} ...
// CODE GENERATED AUTOMATICALLY
// DO NOT EDIT
package {{.PkgName}}

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// loggingMiddleware wraps Service and logs request information to the provided logger
type loggingMiddleware struct {
	logger log.Logger
	svc    {{ .Iface.Name }}
}
{{$methods := .HTTPMethods}}
{{range .Iface.Methods -}}
	{{$method := index $methods .Name}}
	// {{.Name}} ...
	func (s *loggingMiddleware) {{.Name}}({{joinFullVariables .Args ","}}) ({{joinFullVariables .Results ","}}) {
		defer func(begin time.Time) {
			_ = s.wrap(err).Log(
				"method", "{{.Name}}",
				{{$args := popFirst .Args -}}
				{{range $arg := $args -}}
					{{if notin $method.LogIgnores $arg.Name}}"{{$arg.Name}}", {{$arg.Name}},{{end}} 
				{{end -}}
				{{$args := popLast .Results -}}
				{{range $arg := $args -}}
					{{if notin $method.LogIgnores $arg.Name}}"{{$arg.Name}}", {{$arg.Name}},{{end}}
				{{end -}}
				"err", err,
				"elapsed", time.Since(begin),
			)
		}(time.Now())
		return s.svc.{{.Name}}({{joinVariableNamesWithEllipsis .Args ","}})
	}
{{end}}

func (s *loggingMiddleware) wrap(err error) log.Logger {
	lvl := level.Debug
	if err != nil {
		lvl = level.Error
	}
	return lvl(s.logger)
}

// NewLoggingMiddleware ...
func NewLoggingMiddleware(logger log.Logger, svc {{ .Iface.Name }}) {{ .Iface.Name }} {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}
`

// Logging ...
type Logging struct {
	*template.Template
	filePath []string
	imports  imports
}

// Generate ...
func (s *Logging) Generate(info api.Interface) (err error) {
	info.PkgName = path.Base(info.AbsOutputPath)
	info.AbsOutputPath = strings.Join(append(strings.Split(info.AbsOutputPath, "/"), s.filePath...), "/")
	dir, _ := path.Split(info.AbsOutputPath)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return
	}
	file, err := os.Create(info.AbsOutputPath)
	defer func() {
		_ = file.Close()
	}()
	t := template.Must(s.Parse(loggingTpl))
	if err = t.Execute(file, info); err != nil {
		return
	}
	err = s.imports.GoImports(info.AbsOutputPath)
	return
}

// NewLogging ...
func NewLogging(template *template.Template, filePath []string, imports imports) *Logging {
	return &Logging{
		Template: template,
		filePath: filePath,
		imports:  imports,
	}
}
