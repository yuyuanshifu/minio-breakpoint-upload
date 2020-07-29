package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	oplogging "github.com/op/go-logging"
)

const (
	logDir      = "log"
	logSoftLink = "latest_log"
	module      = "minio-breakpoint-upload"
)

var (
	defaultFormatter = `%{time:2006-01-02 - 15:04:05.000} %{longfile} %{color:bold}▶ [%{level:.6s}] %{message}%{color:reset}`
	LOG    *oplogging.Logger
)

type Server struct {
	Log Log
}

type Log struct {
	Prefix  string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	LogFile bool   `mapstructure:"log-file" json:"logFile" yaml:"log-file"`
	Stdout  string `mapstructure:"stdout" json:"stdout" yaml:"stdout"`
	File    string `mapstructure:"file" json:"file" yaml:"file"`
}

func init() {
	c := *new(Log)
	if c.Prefix == "" {
		_ = fmt.Errorf("logger prefix not found")
	}
	c.File = "DEBUG"
	logger := oplogging.MustGetLogger(module)
	var backends []oplogging.Backend
	backends = registerStdout(c, backends)
	backends = registerFile(c, backends)

	oplogging.SetBackend(backends...)
	LOG = logger
}

func registerStdout(c Log, backends []oplogging.Backend) []oplogging.Backend {
	if c.Stdout != "" {
		level, err := oplogging.LogLevel(c.Stdout)
		if err != nil {
			fmt.Println(err)
		}
		backends = append(backends, createBackend(os.Stdout, c, level))
	}

	return backends
}

func registerFile(c Log, backends []oplogging.Backend) []oplogging.Backend {
	if c.File != "" {
		if ok, _ := pathExists(logDir); !ok {
			// directory not exist
			fmt.Println("create log directory")
			_ = os.Mkdir(logDir, os.ModePerm)
		}
		fileWriter, err := rotatelogs.New(
			logDir+string(os.PathSeparator)+"%Y-%m-%d-%H-%M.log",
			// generate soft link, point to latest log file
			rotatelogs.WithLinkName(logSoftLink),
			// maximum time to save log files
			rotatelogs.WithMaxAge(7*24*time.Hour),
			// time period of log file switching
			rotatelogs.WithRotationTime(24*time.Hour),
		)
		if err != nil {
			fmt.Println(err)
			return backends
		}
		level, err := oplogging.LogLevel(c.File)
		if err != nil {
			fmt.Println(err)
		}
		backends = append(backends, createBackend(fileWriter, c, level))
	}

	return backends
}

func createBackend(w io.Writer, c Log, level oplogging.Level) oplogging.Backend {
	backend := oplogging.NewLogBackend(w, c.Prefix, 0)
	stdoutWriter := false
	if w == os.Stdout {
		stdoutWriter = true
	}
	format := getLogFormatter(c, stdoutWriter)
	backendLeveled := oplogging.AddModuleLevel(oplogging.NewBackendFormatter(backend, format))
	backendLeveled.SetLevel(level, module)
	return backendLeveled
}

func getLogFormatter(c Log, stdoutWriter bool) oplogging.Formatter {
	pattern := defaultFormatter
	if !stdoutWriter {
		// Color is only required for console output
		// Other writers don't need %{color} tag
		pattern = strings.Replace(pattern, "%{color:bold}", "", -1)
		pattern = strings.Replace(pattern, "%{color:reset}", "", -1)
	}
	if !c.LogFile {
		// Remove %{logfile} tag
		pattern = strings.Replace(pattern, "%{longfile}", "", -1)
	}
	return oplogging.MustStringFormatter(pattern)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

