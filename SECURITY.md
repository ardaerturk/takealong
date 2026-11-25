# Security

Use Kitwake on a computer you trust. It runs a local process under your user account. The USB carries executables and prompts; authentication stays with the host's Codex configuration unless you change that configuration yourself.

- Never distribute an API key, refresh token, SSH private key, profile with embedded credentials, or chat database on an ordinary USB drive.
- Treat a returned or borrowed drive as untrusted. A modified binary can act as you; a modified recipe can influence what gets sent to an AI provider.
- The read-only sandbox applies to Codex's local tools. It is not a guarantee about connected services, malicious hosts, or the behavior of other agents. Review your MCP servers and app permissions.
- A lost kit does not normally require rotating a provider key. If you unlocked credentials on a compromised machine, revoke affected sessions and rotate credentials from a trusted device.
- Official updates are deliberate downloads, not automatic changes on insertion. The build verifies pinned upstream archives and includes their notices.

Please [report a vulnerability privately](https://github.com/ardaerturk/kitwake/security/advisories/new). Do not put credentials or private session logs in public issues.
