package confparse

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type Key struct {
	sec string
	key string
}

type ConfParse map[Key]string

func New(filename string) (ConfParse, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := make(ConfParse)
	sec := ""

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		if line[0] == '[' && line[len(line)-1] == ']' {
			if sec = strings.TrimSpace(line[1 : len(line)-1]); len(sec) == 0 {
				return nil, errors.New("Section [] Not Allow")
			}
			continue
		}

		fields := strings.Split(line, "=")
		if len(fields) != 2 {
			return nil, errors.New("Config File Error")
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		if len(key) == 0 || len(value) == 0 {
			return nil, errors.New("Key Value Not Valid")
		}

		conf[Key{sec, key}] = value
	}
	return conf, nil
}

func (conf ConfParse) GetStr(sec, key string) (string, error) {
	value, ok := conf[Key{sec, key}]
	if !ok {
		return "", errors.New("Key \"" + key + "\" Not Found")
	}
	return value, nil
}

func (conf ConfParse) GetInt(sec, key string) (int, error) {
	value, ok := conf[Key{sec, key}]
	if !ok {
		return 0, errors.New("Key \"" + key + "\" Not Found")
	}

	iValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("Value is Invalid")
	}
	return iValue, nil
}
