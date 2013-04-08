// Package ini is a parser for ini files implemented using the 'text/scanner' package.
package ini

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/scanner"
)

const (
	tokenSectionStart   = '['
	tokenSectionStop    = ']'
	tokenCommentClassic = ';'
	tokenCommentHash    = '#'
	tokenSpace          = ' '
	tokenLF             = '\n'
	tokenCR             = '\r'
)

type Ini struct {
	data map[string]map[string]string
	rw   sync.RWMutex
}

// Instantiates a new Ini struct
func newIni() *Ini {
	return &Ini{data: make(map[string]map[string]string)}
}

// Get() returns the value associated to section and key. If key is not in a section, use ""
// If key does not exist, Get() returns an empty string.
func (ini *Ini) Get(section, key string) string {
	ini.rw.RLock()
	defer ini.rw.RUnlock()

	if _, ok := ini.data[section]; !ok {
		return ""
	}

	return ini.data[section][key]
}

// Set() sets the value of a key for a given section.
func (ini *Ini) Set(section, key, value string) {
	ini.rw.Lock()
	defer ini.rw.Unlock()

	if _, ok := ini.data[section]; !ok {
		ini.data[section] = make(map[string]string)
	}
	ini.data[section][key] = value
}

// Load() load the ini configuration contained in the Reader until EOF.
func Load(r io.Reader) (*Ini, error) {
	s := new(scanner.Scanner).Init(r)
	s.Mode = scanner.ScanStrings
	s.Whitespace = 1 << '\t'

	ini := newIni()
	currentSection := ""
	for {
		token := s.Peek()
		switch {
		case token == scanner.EOF:
			return ini, nil
		case token == tokenCommentClassic || token == tokenCommentHash:
			readCommentLine(s)
			break
		case token == '\n' || token == '\r':
			s.Scan()
			break
		case token == tokenSectionStart:
			var err error
			currentSection, err = readSection(s)
			if err != nil {
				return nil, err
			}
			break
		default:
			key, err := readKey(s)
			if err != nil {
				return nil, err
			}
			value, err := readValue(s)
			if err != nil {
				return nil, err
			}
			ini.Set(currentSection, key, value)
			break
		}
	}

	panic("unreachable")
}

func readSection(s *scanner.Scanner) (string, error) {
	buffer := new(bytes.Buffer)
	for {
		pos := s.Pos()
		token := s.Scan()
		switch {
		case token == tokenSectionStart:
			break
		case token == tokenSectionStop:
			return buffer.String(), nil
		case token == '\n' || token == '\r':
			return "", fmt.Errorf("While reading a section, got newline. %s", pos.String())
		default:
			buffer.WriteRune(token)
			break
		}
	}
	return buffer.String(), nil
}

func readValue(s *scanner.Scanner) (string, error) {
	buffer := new(bytes.Buffer)
	for {
		token := s.Scan()
		switch {
		case token == scanner.EOF:
			return buffer.String(), nil
		case token == scanner.String:
			value := strings.TrimRight(strings.TrimLeft(s.TokenText(), "\""), "\"")
			return value, nil
		case token == tokenLF:
			return buffer.String(), nil
		case token == tokenCR:
			break
		case token == tokenSpace:
			if buffer.Len() == 0 {
				break
			}
			buffer.WriteRune(token)
		default:
			buffer.WriteRune(token)
		}
	}

	return buffer.String(), nil
}

func readKey(s *scanner.Scanner) (string, error) {
	buffer := new(bytes.Buffer)
	for {
		pos := s.Pos()
		token := s.Scan()
		switch {
		case token == scanner.EOF:
			return "", fmt.Errorf("While reading a key, got EOF. %s", pos.String())
		case token == tokenSpace:
			break
		case token == '=':
			return buffer.String(), nil
		case token == scanner.String:
			return "", fmt.Errorf("While reading a key, got string. %s", pos.String())
		default:
			buffer.WriteRune(token)
		}
	}

	return buffer.String(), nil
}

func readCommentLine(s *scanner.Scanner) {
	for {
		token := s.Scan()
		if token == '\n' {
			return
		}
	}
}
