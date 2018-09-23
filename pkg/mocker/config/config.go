package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// NewFromFilename returns a new config from a filename
func NewFromFilename(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}
	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}
	return NewFromBytes(contents)
}

func NewFromBytes(b []byte) (*Config, error) {
	c := &Config{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}
	return c, nil
}

type Config struct {
	ApiVersion string  `yaml:"apiVersion"`
	Images     []Image `yaml:"images"`
}

type Image struct {
	From   string `yaml:"from"`
	Parent string `yaml:"parent"`

	ExternalFiles []*ExternalFile `yaml:"external"`

	WorkDir string   `yaml:"workdir"`
	Steps   []string `yaml:"steps"`
	Output  []string `yaml:"output"`

	Package *Package `yaml:"package"`
}

type Package struct {
	Repo    []string `yaml:"repo"`
	Gpg     []string `yaml:"gpg"`
	Install []string `yaml:"install"`
}

type ExternalFile struct {
	Source      string `yaml:"src"`
	Destination string `yaml:"dst"`
	Sha256      string `yaml:"sha256"`

	Install []string `yaml:"install"`
}
