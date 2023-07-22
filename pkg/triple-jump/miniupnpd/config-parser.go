package miniupnpd

import (
	"bytes"
	"fmt"
	"github.com/valyala/bytebufferpool"
	"os"
)

type Config []ConfigLine

type ConfigLine struct {
	Key     string
	Value   string
	KVSplit byte
}

func (c Config) String() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	for _, line := range c {
		_, _ = b.WriteString(line.String())
		_ = b.WriteByte('\n')
	}
	return b.String()
}

func (l ConfigLine) String() string {
	return l.Key + string(l.KVSplit) + l.Value
}

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseConfig(b)
}

func SaveConfig(path string, c Config) error {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	for _, line := range c {
		_, _ = b.WriteString(line.String())
		_ = b.WriteByte('\n')
	}

	return os.WriteFile(path, b.Bytes(), 0644)
}

func parseConfig(b []byte) (Config, error) {
	var (
		c   Config
		err error
	)
	for _, lineRaw := range bytes.Split(b, []byte{'\n'}) {
		if len(lineRaw) == 0 || lineRaw[0] == '#' || len(bytes.TrimSpace(lineRaw)) == 0 {
			continue
		}

		var line ConfigLine
		line, err = parseConfigLine(lineRaw)
		if err != nil {
			return nil, err
		}
		c = append(c, line)
	}
	return c, nil
}

func parseConfigLine(b []byte) (ConfigLine, error) {
	var (
		line ConfigLine
	)

	if bytes.HasPrefix(b, []byte("allow")) || bytes.HasPrefix(b, []byte("deny")) {
		line.KVSplit = ' '
	} else {
		line.KVSplit = '='
	}

	res := bytes.SplitN(b, []byte{line.KVSplit}, 2)
	if len(res) != 2 {
		return line, fmt.Errorf("invalid config line: %s", b)
	}

	line.Key = string(bytes.TrimSpace(res[0]))
	line.Value = string(bytes.TrimSpace(res[1]))

	return line, nil
}
