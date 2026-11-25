package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandLine(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
		bad  bool
	}{
		{nil, "open", false}, {[]string{"--profile", "work", "brief", "sample"}, "brief", false},
		{[]string{"--agent", "claude", "open"}, "open", false}, {[]string{"--agent", "claude", "brief"}, "", true},
		{[]string{"run"}, "", true}, {[]string{"--installed", "remote", "studio"}, "", true},
		{[]string{"--profile"}, "", true}, {[]string{"brieff"}, "", true},
	} {
		r, e := parse(c.args)
		if (e != nil) != c.bad || (!c.bad && r.verb != c.want) {
			t.Errorf("%v: %v, %v", c.args, r, e)
		}
	}
}
func TestRemoteKeepsUserInputOutOfRemoteShell(t *testing.T) {
	const id = "b92f8600-465e-4a9e-8c7d-f2dd1b36bb25"
	for _, x := range [][2]string{{"-o", ""}, {"studio;other", ""}, {"studio", "abc;exit"}} {
		if _, e := remoteArgs(x[0], x[1]); e == nil {
			t.Fatal(x)
		}
	}
	a, e := remoteArgs("studio", id)
	if e != nil || a[len(a)-1] != "exec codex resume "+id || !strings.Contains(strings.Join(a, " "), "StrictHostKeyChecking=yes") {
		t.Fatalf("%v %v", a, e)
	}
}
func TestRecipesAreSmallDataFiles(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "recipes")
	if e := os.Mkdir(d, 0700); e != nil {
		t.Fatal(e)
	}
	write := func(name, contents string) {
		t.Helper()
		if e := os.WriteFile(filepath.Join(d, name+".md"), []byte(contents), 0600); e != nil {
			t.Fatal(e)
		}
	}
	original, e := recipe(root, "brief")
	if e != nil || original == "" {
		t.Fatal(e)
	}
	write("brief", "Use line one.\nThen line two.")
	s, e := recipe(root, "brief")
	if e != nil || s != "Use line one.\nThen line two." {
		t.Fatal(s, e)
	}
	for _, x := range []string{"../private", "missing", "../recipes/brief"} {
		if _, e := recipe(root, x); e == nil {
			t.Fatal(x)
		}
	}
	write("big", strings.Repeat("x", 16385))
	write("empty", " \n")
	for _, n := range []string{"big", "empty"} {
		if _, e := recipe(root, n); e == nil {
			t.Fatal(n)
		}
	}
	if os.Symlink(filepath.Join(d, "brief.md"), filepath.Join(d, "linked.md")) == nil {
		if _, e := recipe(root, "linked"); e == nil {
			t.Fatal("followed a symlink")
		}
	}
}
func TestProjectIsNeverTheDriveByAccident(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Chdir(root)
	realRoot, e := filepath.EvalSymlinks(root)
	if e != nil {
		t.Fatal(e)
	}
	got, e := project(realRoot, "")
	if e != nil || got != filepath.Join(home, "Kitwake Workspace") {
		t.Fatal(got, e)
	}
	folder := filepath.Join(root, "A project")
	if e := os.Mkdir(folder, 0700); e != nil {
		t.Fatal(e)
	}
	got, e = project(realRoot, folder)
	if e != nil || got != folder {
		t.Fatal(got, e)
	}
}
