<p align="center"><img src="docs/assets/mark.svg" alt="" width="80" height="80"></p>
<h1 align="center">Kitwake</h1>
<p align="center">A coding agent you can carry on a USB drive.</p>
<p align="center"><a href="https://github.com/ardaerturk/kitwake/releases/latest/download/kitwake-usb.zip"><strong>Download the USB kit</strong></a> · <a href="https://github.com/ardaerturk/kitwake/releases/latest">Release notes &amp; checksum</a></p>

Kitwake opens Codex in a terminal on a trusted Linux or Windows computer. The kit includes the CLI. You don't need to install Codex, Node, or npm first.

## Start here

1. Download the kit, unzip it, and copy the **Kitwake** folder to a USB drive.
2. On **Linux**, open that folder in a terminal and run `./kitwake`. On **Windows**, double-click **kitwake.exe**.
3. Follow Codex's sign-in prompt, or use a provider already configured on that computer.

The USB carries the program and your recipes. Your login, files, and sessions stay on the computer you use. Starting requires one deliberate action; insertion alone does nothing. Cloud models need a connection and an authorized account.

## Three things to try

| I want to… | Run |
| --- | --- |
| Open Codex in a project | `./kitwake open ~/projects/notes` |
| Get a short, read-only briefing | `./kitwake brief ~/projects/notes` |
| Continue a previous session | `./kitwake resume` |

On Windows, use `.\kitwake.exe` and a Windows folder path. To continue on a different computer **without copying the session**, use [a trusted SSH host](docs/guide.md#continue-on-another-computer).

## Make it yours

**Start with the same task.** Put a file named `Start.md` next to the launcher. Write what you want Codex to do, such as “Tell me what changed in this folder and what to read first.” The next time you open Kitwake, it starts a fresh chat with that message and leaves the terminal open for follow-up questions. Run `./kitwake open` to skip it.

**Keep several tasks.** Add a plain text file called `recipes/review.md` to the kit:

```text
Review the changes in this folder. Name the most important risk
and one test that would give me confidence. Do not edit files.
```

Run `./kitwake run review ~/projects/notes`. A recipe becomes a Codex prompt, never a shell script. Recipe runs request the read-only sandbox and disable approval escalation. Your configured provider still receives the prompt and any context the agent sends. Review connected tools before using sensitive files.

**Supported:** bundled Codex for x64 Linux and Windows. Existing Codex, Claude Code, and OpenCode can also be opened explicitly. macOS and ARM do not have bundled runtimes in this release.

[User guide](docs/guide.md) · [Build and contribute](CONTRIBUTING.md) · [Related projects](docs/related.md) · [Security](SECURITY.md) · [MIT license](LICENSE)
