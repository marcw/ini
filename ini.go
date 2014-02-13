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

// Ini structure contains the data and a RWMutex for concurrency safety
type Ini struct {
	data map[string]map[string]string
	rw   sync.RWMutex
}

// Instantiates a new Ini structure
func NewIni() *Ini {
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

	ini.set(section, key, value)
}

func (ini *Ini) Has(section, key string) bool {
	ini.rw.RLock()
	defer ini.rw.RUnlock()

	if _, ok := ini.data[section]; !ok {
		return false
	}
	if _, ok := ini.data[section][key]; !ok {
		return false
	}
	return true
}

// Unsafe version of Set
func (ini *Ini) set(section, key, value string) {
	if _, ok := ini.data[section]; !ok {
		ini.data[section] = make(map[string]string)
	}
	ini.data[section][key] = value
}

// ReadFrom() read the ini configuration contained in the Reader r until EOF.
func (ini *Ini) ReadFrom(r io.Reader) (int64, error) {
	ini.rw.Lock()
	defer ini.rw.Unlock()

	s := new(scanner.Scanner).Init(r)
	s.Mode = scanner.ScanStrings
	s.Whitespace = 1 << '\t'

	currentSection := ""
	for {
		token := s.Peek()
		switch {
		case token == scanner.EOF:
			return 0, nil
		case token == tokenCommentClassic || token == tokenCommentHash:
			ini.readCommentLine(s)
			break
		case token == '\n' || token == '\r':
			s.Scan()
			break
		case token == tokenSectionStart:
			var err error
			currentSection, err = ini.readSection(s)
			if err != nil {
				return -1, err
			}
			break
		default:
			key, err := ini.readKey(s)
			if err != nil {
				return -1, err
			}
			value, err := ini.readValue(s)
			if err != nil {
				return -1, err
			}
			ini.set(currentSection, key, value)
			break
		}
	}

	panic("unreachable")
}

// WriteTo() writes the configuration in an ini format to the Writer writer.
func (ini *Ini) WriteTo(writer io.Writer) (int64, error) {
	ini.rw.RLock()
	defer ini.rw.RUnlock()
	var nw int64

	// Starting with the "" section

	if data, ok := ini.data[""]; ok {
		for k := range data {
			n, err := fmt.Fprintf(writer, "%s=%q\n", k, data[k])
			nw = nw + int64(n)
			if err != nil {
				return nw, err
			}
		}
	}

	for section := range ini.data {
		if section == "" {
			continue
		}
		index := 0
		for k := range ini.data[section] {
			if index == 0 {
				n, err := fmt.Fprintf(writer, "[%s]\n", section)
				nw = nw + int64(n)
				if err != nil {
					return nw, err
				}
			}
			n, err := fmt.Fprintf(writer, "%s=%q\n", k, ini.data[section][k])
			nw = nw + int64(n)
			if err != nil {
				return nw, err
			}
			index++
		}
	}
	return nw, nil
}

func (ini *Ini) readSection(s *scanner.Scanner) (string, error) {
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

func (ini *Ini) readValue(s *scanner.Scanner) (string, error) {
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

func (ini *Ini) readKey(s *scanner.Scanner) (string, error) {
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

func (ini *Ini) readCommentLine(s *scanner.Scanner) {
	for {
		token := s.Scan()
		if token == '\n' {
			return
		}
	}
}
