package lib

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigProc struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Cwd  string   `yaml:"cwd"`
	Args []string `yaml:"args"`
}

type Config struct {
	Procs []ConfigProc      `yaml:"procs"`
	Env   map[string]string `yaml:"env"`
}

func ReadConfig(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	duplicateProcessNames := getDuplicateEntries(c.Procs, func(proc ConfigProc) string { return proc.Name })

	if len(duplicateProcessNames) > 0 {
		return nil, fmt.Errorf("duplicate process names found: %s", strings.Join(duplicateProcessNames, ", "))
	}

	return c, nil
}

func getDuplicateEntries[K interface{}, L comparable](list []K, mapper func(v K) L) []L {
	result := make([]L, 0)

	workspace := make(map[L]bool, 0)

	for _, v := range list {
		key := mapper(v)
		if visited := workspace[key]; visited {
			result = append(result, key)
		}
		workspace[key] = true
	}

	return result
}
