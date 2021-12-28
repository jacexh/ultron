package ultron

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jacexh/multiconfig"
)

type (
	Option struct {
		Server ServerOption
		Logger LoggerOption
	}

	ServerOption struct {
		HTTPAddr string `default:":2017" yaml:"http_addr,omitempty" json:"http_addr,omitempty" toml:"http_addr"`
		GRPCAddr string `default:":2021" yaml:"grpc_addr,omitempty" json:"grpc_addr,omitempty" toml:"grpc_addr"`
	}

	LoggerOption struct {
		Level      string `default:"info" yaml:"level,omitempty" json:"level,omitempty" toml:"level"`
		FileName   string `yaml:"filename,omitempty" json:"filename,omitempty" toml:"filename"`
		MaxSize    int    `default:"100" yaml:"max_size,omitempty" json:"max_size,omitempty" toml:"max_size"`
		MaxBackups int    `default:"30" yaml:"max_backups,omitempty" json:"max_backups,omitempty" toml:"max_backups"`
	}
)

var (
	configFileBaseName          = "config"
	configFileFormat            = "yml"
	searchInPaths      []string = []string{"."}
)

var (
	loadedOption *Option
)

func findInDir(dir string, file string) string {
	fp := filepath.Join(dir, file)
	fi, err := os.Stat(fp)
	if err == nil && !fi.IsDir() {
		return fp
	}
	return ""
}

func findConfigFile() string {
	fp := fmt.Sprintf("%s.%s", configFileBaseName, configFileFormat)
	for _, dir := range searchInPaths {
		if path := findInDir(dir, fp); path != "" {
			return path
		}
	}
	return ""
}

func loadConfig() *Option {
	f := findConfigFile()
	opt := new(Option)
	loader := multiconfig.NewWithPathAndEnvPrefix(f, "ULTRON")
	loader.MustLoad(opt)

	buildLogger(opt.Logger) // todo
	return opt
}

func init() {
	loadedOption = loadConfig()
}
