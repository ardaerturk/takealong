// Kitwake opens a coding agent from a portable drive.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "0.1.0"

const usage = `Kitwake · A coding agent you can carry.

  kitwake [folder]                Open Codex
  kitwake brief [folder]          Brief a folder without editing it
  kitwake resume [session-id]     Continue a session on this computer
  kitwake remote host [id]        Continue on a trusted SSH host

  kitwake run name [folder]       Run a recipe from the drive
  kitwake agents                  See which agents are available
  kitwake login                   Sign in to Codex with a device code
  kitwake doctor                  Check the selected runtime

Options (before the command):
  --installed        Use Codex from PATH instead of the drive
  --profile name     Use a Codex profile configured on this computer
  --agent name       Open codex, claude, or opencode
  --version          Show the Kitwake version
  --help             Show this guide`

type options struct {
	installed      bool
	profile, agent string
}
type request struct {
	opts options
	verb string
	args []string
}

var commands = map[string]int{"open": 1, "brief": 1, "run": 2, "resume": 1, "remote": 2, "agents": 0, "doctor": 0, "login": 0, "help": 0, "version": 0}

func parse(args []string) (request, error) {
	r := request{opts: options{agent: "codex"}, verb: "open"}
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		flag := args[0]
		args = args[1:]
		switch flag {
		case "--help", "-h":
			r.verb = "help"
			return r, nil
		case "--version":
			r.verb = "version"
			return r, nil
		case "--installed":
			r.opts.installed = true
		case "--profile", "--agent":
			if len(args) == 0 || !simpleName.MatchString(args[0]) {
				return r, fmt.Errorf("%s needs a name", flag)
			}
			if flag == "--profile" {
				r.opts.profile = args[0]
			} else {
				r.opts.agent = args[0]
			}
			args = args[1:]
		default:
			return r, fmt.Errorf("unknown option %q; run kitwake --help", flag)
		}
	}
	if len(args) > 0 {
		if _, ok := commands[args[0]]; ok {
			r.verb = args[0]
			args = args[1:]
		} else if st, err := os.Stat(args[0]); err != nil || !st.IsDir() {
			return r, fmt.Errorf("%q is not a command or an existing folder; run kitwake --help", args[0])
		}
	}
	r.args = args
	if len(args) > commands[r.verb] {
		return r, fmt.Errorf("too many arguments for %s", r.verb)
	}
	if (r.verb == "run" || r.verb == "remote") && len(args) == 0 {
		return r, fmt.Errorf("%s needs a name; run kitwake --help", r.verb)
	}
	if r.opts.agent != "codex" && r.opts.agent != "claude" && r.opts.agent != "opencode" {
		return r, fmt.Errorf("unknown agent %q; choose codex, claude, or opencode", r.opts.agent)
	}
	if r.opts.agent != "codex" && (r.verb != "open" || r.opts.profile != "" || r.opts.installed) {
		return r, errors.New("use --agent with open; profiles, recipes, and sessions use Codex")
	}
	if (r.verb == "remote" || r.verb == "agents") && (r.opts.profile != "" || r.opts.installed) {
		return r, fmt.Errorf("%s does not use a local Codex profile or runtime", r.verb)
	}
	return r, nil
}

func main() {
	code := run(os.Args[1:])
	os.Exit(code)
}

func run(args []string) int {
	r, err := parse(args)
	if err == nil {
		switch r.verb {
		case "help":
			fmt.Println(usage)
			return 0
		case "version":
			fmt.Println("Kitwake " + version)
			return 0
		}
		var self string
		self, err = os.Executable()
		if err == nil {
			var real string
			real, err = filepath.EvalSymlinks(self)
			if err == nil {
				err = launch(filepath.Dir(real), r)
			}
		}
	}
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		if n := exit.ExitCode(); n > 0 {
			return n
		}
		return 1
	}
	fmt.Fprintln(os.Stderr, "Kitwake:", err)
	return 1
}

func hostBinary(name string) (string, error) {
	if runtime.GOOS == "windows" {
		name += ".exe"
	} // Never pass filenames or prompts through cmd.exe.
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(p) {
		return "", errors.New("expected an absolute executable path")
	}
	return p, nil
}
