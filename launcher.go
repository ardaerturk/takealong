package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

//go:embed recipes/brief.md
var defaultBrief string
var simpleName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,47}$`)
var hostName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,126}$`)
var sessionID = regexp.MustCompile(`(?i)^[a-f0-9]{8}(-[a-f0-9]{4}){3}-[a-f0-9]{12}$`)

type engine struct{ path, source string }

func bundled(root string) string {
	if runtime.GOARCH != "amd64" || (runtime.GOOS != "linux" && runtime.GOOS != "windows") {
		return ""
	}
	n := "codex"
	if runtime.GOOS == "windows" {
		n += ".exe"
	}
	return filepath.Join(root, "runtime", runtime.GOOS+"-x64", "bin", n)
}
func resolve(root string, o options) (engine, error) {
	if !o.installed && o.agent == "codex" {
		if p := bundled(root); p != "" {
			if f, e := os.Stat(p); e == nil && f.Mode().IsRegular() {
				return engine{p, "USB bundle"}, nil
			}
		}
	}
	if p, e := hostBinary(o.agent); e == nil {
		return engine{p, "host PATH"}, nil
	}
	if o.agent == "codex" && !o.installed {
		return engine{}, errors.New("Codex was not found. Extract the full USB kit, including its runtime folder, on x64 Linux or Windows")
	}
	return engine{}, fmt.Errorf("%s was not found on PATH. Open the bundled Codex with takealong", o.agent)
}

func project(root, explicit string) (string, error) {
	p := explicit
	if p == "" {
		cwd, e := os.Getwd()
		if e != nil {
			return "", e
		}
		real, e := filepath.EvalSymlinks(cwd)
		if e != nil {
			return "", e
		}
		rel, e := filepath.Rel(root, real)
		if e == nil && filepath.IsLocal(rel) {
			home, e := os.UserHomeDir()
			if e != nil {
				return "", e
			}
			p = filepath.Join(home, "Takealong Workspace")
			// Keep existing users in their original workspace without moving any files.
			if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) {
				legacy := filepath.Join(home, "Kitwake Workspace")
				if info, err := os.Lstat(legacy); err == nil && info.IsDir() {
					p = legacy
				}
			}
			if e = os.MkdirAll(p, 0700); e != nil {
				return "", e
			}
		} else {
			p = cwd
		}
	}
	p, e := filepath.Abs(p)
	if e != nil {
		return "", e
	}
	f, e := os.Stat(p)
	if e != nil || !f.IsDir() {
		return "", fmt.Errorf("folder not found: %s", p)
	}
	return p, nil
}

func recipe(root, name string) (string, error) {
	if !simpleName.MatchString(name) {
		return "", errors.New("use a recipe name, not a path")
	}
	dir := filepath.Join(root, "recipes")
	if st, e := os.Lstat(dir); e == nil && !st.IsDir() {
		return "", errors.New("recipes must be a real directory")
	}
	p := filepath.Join(dir, name+".md")
	text, e := readPrompt(p)
	if errors.Is(e, os.ErrNotExist) && name == "brief" {
		return strings.TrimSpace(defaultBrief), nil
	}
	if errors.Is(e, os.ErrNotExist) {
		return "", fmt.Errorf("recipe %q not found; add recipes/%s.md to the drive", name, name)
	}
	return text, e
}

func readPrompt(p string) (string, error) {
	st, e := os.Lstat(p)
	if e != nil {
		return "", e
	}
	if !st.Mode().IsRegular() || st.Size() > 16384 {
		return "", errors.New("use a regular text file of at most 16 KiB")
	}
	f, e := os.Open(p)
	if e != nil {
		return "", e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, 16385))
	if e != nil {
		return "", e
	}
	s := strings.TrimSpace(string(b))
	if s == "" || len(b) > 16384 || !utf8.Valid(b) || strings.ContainsRune(s, 0) {
		return "", errors.New("enter plain UTF-8 text of at most 16 KiB")
	}
	return s, nil
}

// Start.md may provide a first message. It cannot select an executable, folder or provider.
func openingArgs(root, folder string, useStart bool) ([]string, error) {
	args := []string{"--cd", folder, "--sandbox", "read-only"}
	if !useStart {
		return args, nil
	}
	text, err := readPrompt(filepath.Join(root, "Start.md"))
	if errors.Is(err, os.ErrNotExist) {
		return args, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Start.md: %w; use takealong open for a blank session", err)
	}
	return append(args, "--ask-for-approval", "never", "--", text), nil
}

func remoteArgs(host, id string) ([]string, error) {
	if !hostName.MatchString(host) {
		return nil, errors.New("use an SSH host alias, for example: takealong remote studio")
	}
	line := "exec codex resume --all"
	if id != "" {
		if !sessionID.MatchString(id) {
			return nil, errors.New("expected a session UUID")
		}
		line = "exec codex resume " + id
	}
	return []string{"-t", "-o", "StrictHostKeyChecking=yes", "-o", "ForwardAgent=no", "-o", "ForwardX11=no", "-o", "ClearAllForwardings=yes", host, line}, nil
}
func at(v []string, i int) string {
	if i >= len(v) {
		return ""
	}
	return v[i]
}

func launch(root string, r request) error {
	if r.verb == "remote" {
		a, e := remoteArgs(r.args[0], at(r.args, 1))
		if e != nil {
			return e
		}
		p, e := hostBinary("ssh")
		if e != nil {
			return errors.New("remote needs an OpenSSH client on this computer")
		}
		return execute(p, a, "")
	}
	if r.verb == "agents" {
		for _, name := range []string{"codex", "claude", "opencode"} {
			eng, e := resolve(root, options{agent: name})
			if e != nil {
				fmt.Printf("%-10s not found\n", name)
			} else {
				fmt.Printf("%-10s %s\n", name, eng.source)
			}
		}
		return nil
	}
	eng, e := resolve(root, r.opts)
	if e != nil {
		return e
	}
	if r.verb == "doctor" {
		out, e := exec.Command(eng.path, "--version").CombinedOutput()
		if e != nil {
			return fmt.Errorf("could not start %s: %w", eng.path, e)
		}
		fmt.Printf("Takealong  %s\nAgent    %s (%s)\nSystem   %s/%s\n", version, strings.TrimSpace(string(out)), eng.source, runtime.GOOS, runtime.GOARCH)
		return nil
	}
	if r.opts.agent != "codex" {
		p, e := project(root, at(r.args, 0))
		if e != nil {
			return e
		}
		return execute(eng.path, nil, p)
	}
	var a []string
	if eng.source == "USB bundle" || exec.Command(eng.path, "--no-daemon", "--version").Run() == nil {
		a = append(a, "--no-daemon")
	}
	if r.opts.profile != "" {
		a = append(a, "--profile", r.opts.profile)
	}
	switch r.verb {
	case "login":
		a = append(a, "login", "--device-auth")
	case "resume":
		a = append(a, "--sandbox", "read-only", "resume")
		if id := at(r.args, 0); id != "" {
			if !sessionID.MatchString(id) {
				return errors.New("expected a session UUID; omit it to open the picker")
			}
			a = append(a, id)
		} else {
			a = append(a, "--all")
		}
	case "open":
		p, e := project(root, at(r.args, 0))
		if e != nil {
			return e
		}
		opening, err := openingArgs(root, p, r.start)
		if err != nil {
			return err
		}
		a = append(a, opening...)
	case "brief", "run":
		name, i := "brief", 0
		if r.verb == "run" {
			name = r.args[0]
			i = 1
		}
		text, e := recipe(root, name)
		if e != nil {
			return e
		}
		p, e := project(root, at(r.args, i))
		if e != nil {
			return e
		}
		a = append(a, "--ask-for-approval", "never", "exec", "--skip-git-repo-check", "--cd", p, "--sandbox", "read-only", "--", text)
	}
	return execute(eng.path, a, "")
}
