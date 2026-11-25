# Contributing

Kitwake is a small Go launcher. There are no third-party Go dependencies. You only need Go 1.25 or later to work on it.

```sh
go test ./...
go run . --help
go run . --installed doctor
```

The last command uses an installed Codex. The first two don't need an agent or credentials. `main.go` parses commands, `launcher.go` selects and starts an agent, and `recipes/` holds the default prompts. Keep prompts, paths and settings as data; don't add a shell between Kitwake and a local agent.

**Build the USB release.** On Linux or macOS, install Go and Python 3.11+, then run `python3 tools/package.py`. It fetches the pinned upstream distributions, verifies their checksums, cross-compiles both launchers, and writes `dist/kitwake-usb.zip` and `dist/SHA256SUMS`. Use `--target linux --no-zip` for a smaller development kit. Runtime archives are cached outside Git in `.cache/`. No provider account is needed.

**Test a full kit.** After packaging for your platform, run `python3 tools/smoke.py dist/Kitwake`. The test starts a loopback model stub with a temporary Codex home. It checks that the real agent receives the recipe and saves a session. It does not use a live model account.

**Change a runtime.** Update the version and both sets of digests in `tools/runtimes.json` from the official OpenAI release; review upstream notices before redistributing. Never derive the expected archive hash from an unverified download. Test on native Ubuntu and Windows.

**Publish.** Update the Kitwake version and changelog, merge a green commit, and tag it `vX.Y.Z`. The release workflow tests both platforms and builds the downloadable kit before publishing it. Use GitHub issues for bugs and small design proposals. Include the OS, Kitwake version, expected result, and actual result. Redact personal paths and credentials.
