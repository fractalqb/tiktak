package cmd

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode"
)

func CheckTaskName(n string) error {
	if n == "" {
		return errors.New("empty task name")
	}
	for _, r := range n {
		if r == '/' {
			return fmt.Errorf("path separator '/' in task name '%s'", n)
		}
		if unicode.IsSpace(r) {
			return fmt.Errorf("task name '%s' contains space rune", n)
		}
	}
	return nil
}

func CheckPath(p ...string) error {
	for _, n := range p {
		if err := CheckTaskName(n); err != nil {
			return err
		}
	}
	return nil
}

func CheckPathString(p string) error {
	if p == "/" {
		return nil
	}
	if path.IsAbs(p) {
		p = p[1:]
	}
	p = strings.TrimSuffix(p, "/")
	return CheckPath(strings.Split(p, "/")...)
}
