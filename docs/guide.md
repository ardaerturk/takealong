# Using Kitwake

**Open** starts an interactive Codex terminal. **Brief** sends a single request and prints the answer. **Resume** opens Codex's session picker. Opening a folder or reading it may send context to your provider; use a folder you intend to share.

## Pick a folder

Pass a project explicitly, or start Kitwake from the project you want to use. If you launch from the kit itself, Kitwake creates **Kitwake Workspace** in your home directory. It won't choose the USB as your project by accident.

```sh
./kitwake open "/path/to/my project"
./kitwake brief "/path/to/my project"
./kitwake run review "/path/to/my project"
```

On Windows use `.\kitwake.exe` in PowerShell. Launching Kitwake from a source checkout without a bundled runtime uses a working Codex on PATH. A downloaded kit prefers its bundled Codex. To choose your existing CLI, put `--installed` before the command.

## Accounts and providers

A new trusted machine needs its own authorization. Run `./kitwake login` to use Codex's device-code flow if your account allows it. Otherwise use normal interactive sign-in. Removing the USB doesn't sign you out.

Use an existing Codex provider profile with `./kitwake --profile work brief "/path/to/project"`. For Azure, set up the provider and deployment in **this computer's** Codex configuration and supply its credential through your normal secret manager. Keep resource keys, credential helpers, and the Codex home directory off the USB. Rotating a provider key should not require rebuilding the drive. See the [Codex configuration reference](https://developers.openai.com/codex/config-reference/) and [authentication guide](https://developers.openai.com/codex/auth/).

## Continue on another computer

A local session belongs to the Codex environment that created it. To keep its files, history, and permissions together, run the agent there. Set up an OpenSSH host alias such as `studio` on the client. Connect to that host normally once, verify its fingerprint, and ensure `ssh studio 'codex --version'` works.

```sh
./kitwake remote studio
./kitwake remote studio <session-uuid>
```

Remote mode needs OpenSSH on the client, a reachable host, authorized SSH access, and Codex on the host's non-interactive PATH. Kitwake requires a known host key, disables forwarding, and opens the remote session picker unless you supply a UUID. It doesn't upload the project, credentials or recipes. A missing host never triggers a local run.

To resume a local session by ID: `./kitwake resume <session-uuid>`. These are Codex sessions, not ChatGPT website conversations.

## Use another agent

```sh
./kitwake agents
./kitwake --agent claude open "/path/to/project"
./kitwake --agent opencode open "/path/to/project"
```

These two adapters open an existing executable with its own settings and permissions. On Windows, discovery checks native executables on PATH. Only Codex is included in the kit and supports Kitwake recipes. Agents are never switched silently.

## Troubleshooting

| What happened | What to do |
| --- | --- |
| Unsure which runtime will run | Use `./kitwake doctor`. It shows the version, source, and platform without sending a prompt. |
| Permission denied on Linux | Use a location where you're permitted to run software. If your own extraction tool dropped the execute bit, restore it on the Kitwake binary and the runtime's executables, or extract the original archive on Linux with executable permissions preserved. Follow managed-device policy. |
| Wrong architecture | The download supports 64-bit Intel/AMD Linux and Windows. Build Kitwake from source and use an installed CLI for other supported native environments. |
| Can't find a previous session | Use the original user, Codex profile, and computer, or connect with `remote`. No history is stored on the stick by default. |
| Unsure a download is intact | Compare the ZIP's SHA-256 with the release's SHA256SUMS using `sha256sum`, `shasum -a 256`, or PowerShell's `Get-FileHash`. Use a checksum obtained from a trusted source. |

Kitwake ships unsigned executables. Checksums detect corruption; they do not make an untrusted drive or host safe. There is no local model, enrollment service, or automatic USB insertion listener.
