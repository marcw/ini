package ini

import (
	"bytes"
	"testing"
)

func TestIniSetGet(t *testing.T) {
	ini := newIni()
	if ini.Get("", "foo") != "" {
		t.Error("Unknown key should return an empty string")
	}

	if ini.Get("bar", "foo") != "" {
		t.Error("Unknown key in unknown section return an empty string")
	}

	ini.Set("", "foo", "bar")
	if ini.Get("", "foo") != "bar" {
		t.Error("Known key should return correct string")
	}
}

func TestIniLoadMostBasicConfiguration(t *testing.T) {
	config := bytes.NewBufferString("foo=bar")
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("", "foo"); v != "bar" {
		t.Errorf("Expected \"bar\", got %#v", v)
	}
}

func TestIniLoadKeyValueWithString(t *testing.T) {
	config := bytes.NewBufferString(`foo="bar"`)
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("", "foo"); v != "bar" {
		t.Errorf("Expected \"bar\", got %v", v)
	}
}

func TestIniLoadMultiLineKeyValue(t *testing.T) {
	config := bytes.NewBufferString("foo=32\nbar=54")
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("", "foo"); v != "32" {
		t.Errorf("Expected \"32\", got %#v", v)
	}
	if v := ini.Get("", "bar"); v != "54" {
		t.Errorf("Expected \"54\", got %#v", v)
	}
}

func TestIniLoadMultiLineAndComment(t *testing.T) {
	config := bytes.NewBufferString("foo=32.452\n; hihihihi\n# hohohoho\nbar=54")
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("", "foo"); v != "32.452" {
		t.Errorf("Expected \"32\", got %#v", v)
	}
	if v := ini.Get("", "bar"); v != "54" {
		t.Errorf("Expected \"54\", got %#v", v)
	}
}

func TestIniLoadSections(t *testing.T) {
	config := bytes.NewBufferString("foo=32.452\n; hihihihi\n[foobar]\nbar=54")
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("", "foo"); v != "32.452" {
		t.Errorf("Expected \"32\", got %#v", v)
	}
	if v := ini.Get("foobar", "bar"); v != "54" {
		t.Errorf("Expected \"54\", got %#v", v)
	}
}

func TestIniLoadGitConfig(t *testing.T) {
	config := bytes.NewBufferString(
		`
[user]
  name  = Marc Weistroff
  email = marc@example.org
  #email = marc@example.net
[core]
  excludesfile="~/.gitignore"
[alias]
  sdi  = diff --staged
  sdiff = diff --staged
  st   = status
  cat  = cat-file -p
  lg   = log --graph --pretty=tformat:'%Cred%h%Creset -%C(yellow)%d%Creset%s %Cgreen(%an %cr)%Creset' --abbrev-commit --date=relative
  lga  = "!sh -c 'git log --author=\"$1\" -p $2' -"
  lint = "!sh -c 'git status | awk \"/modified/ {print \\$3} /new file/ {print \\$4}\" | xargs -L 1 php -l'"
  uncommit= reset --soft HEAD^
[color]
  branch = auto
  diff = auto
  interactive = auto
  status = auto
[ghi]
    token = 4d3cf26439283fake6fd7ef50c8c6e3c
`)

	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("user", "name"); v != "Marc Weistroff" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("user", "email"); v != "marc@example.org" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("core", "excludesfile"); v != "~/.gitignore" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("alias", "sdi"); v != "diff --staged" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("alias", "lg"); v != "log --graph --pretty=tformat:'%Cred%h%Creset -%C(yellow)%d%Creset%s %Cgreen(%an %cr)%Creset' --abbrev-commit --date=relative" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("alias", "lga"); v != "!sh -c 'git log --author=\\\"$1\\\" -p $2' -" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("alias", "lint"); v != "!sh -c 'git status | awk \\\"/modified/ {print \\\\$3} /new file/ {print \\\\$4}\\\" | xargs -L 1 php -l'" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("alias", "uncommit"); v != "reset --soft HEAD^" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("color", "branch"); v != "auto" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("ghi", "token"); v != "4d3cf26439283fake6fd7ef50c8c6e3c" {
		t.Errorf("Got %#v", v)
	}
}

func TestLoadPHPIni(t *testing.T) {
	config := bytes.NewBufferString(
		`
[PHP]

;;;;;;;;;;;;;;;;;;;
; About php.ini   ;
;;;;;;;;;;;;;;;;;;;

engine = On
short_open_tag = Off
unserialize_callback_func =
error_log = /usr/local/var/log/php-error.log
[CLI Server]
cli_server.color = On
`)
	ini := newIni()
	_, err := ini.ReadFrom(config)
	if err != nil {
		t.Error(err)
	}
	if v := ini.Get("PHP", "engine"); v != "On" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("PHP", "short_open_tag"); v != "Off" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("PHP", "unserialize_callback_func"); v != "" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("PHP", "error_log"); v != "/usr/local/var/log/php-error.log" {
		t.Errorf("Got %#v", v)
	}
	if v := ini.Get("CLI Server", "cli_server.color"); v != "On" {
		t.Errorf("Got %#v", v)
	}
}

func TestWriteTo(t *testing.T) {
	config := bytes.NewBufferString(
		`
foobar="absolute foobaritude"
[PHP]

;;;;;;;;;;;;;;;;;;;
; About php.ini   ;
;;;;;;;;;;;;;;;;;;;

engine = On
short_open_tag = Off
unserialize_callback_func =
error_log = /usr/local/var/log/php-error.log
[CLI Server]
cli_server.color = On
`)
	ini := newIni()
	if _, err := ini.ReadFrom(config); err != nil {
		t.Errorf(err.Error())
	}
	buffer := new(bytes.Buffer)
	if _, err := ini.WriteTo(buffer); err != nil {
		t.Errorf(err.Error())
	}

	ini2 := newIni()
	if _, err := ini2.ReadFrom(buffer); err != nil {
		t.Errorf(err.Error())
	}
	if ini2.Get("", "foobar") != "absolute foobaritude" {
		t.Errorf(ini2.Get("", "foobar"))
	}
	if ini2.Get("PHP", "engine") != "On" {
		t.Errorf(ini2.Get("PHP", "engine"))
	}
	if ini2.Get("CLI Server", "cli_server.color") != "On" {
		t.Errorf(ini2.Get("CLI Server", "cli_server.color"))
	}
}
