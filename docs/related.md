# Related projects

Several projects already put agents on USB drives. Choose based on what you want to carry.

| If you want to carry… | Look at | How it differs from Kitwake |
| --- | --- | --- |
| Your Codex login | [Portable Codex USB](https://github.com/clarenzmauro/portable-codex) | Can keep encrypted authentication on the drive. Kitwake asks the host to authorize the agent. |
| A full personal setup | [AgentKey](https://github.com/davidpengrj/agentkey) | Carries an encrypted provider setup, skills, and memory for Mac and Linux. Kitwake carries prompts and leaves provider settings on the host. |
| A portable development environment | [PortableAgents](https://github.com/ppisljar/PortableAgents) | Packs Codex, Claude Code, and developer tools. Kitwake bundles only Codex and can open a supported agent already installed on the host. |
| A model that works offline | [code-stick](https://github.com/MuhammadUsmanGM/code-stick) | Brings a local model and inference runtime. The default Kitwake download uses a cloud provider. |

Kitwake is for someone who wants one terminal, one familiar first task, and no agent installation on the destination computer. For a session with real files and history, use the computer that already owns it, directly or over SSH.

**Why open the launcher yourself?** Insertion is a poor signal of intent. [Windows restricts AutoRun on USB storage](https://www.microsoft.com/en-us/msrc/blog/2009/04/autorun-changes-in-windows-7/), and [GNOME documents a confirmation before running software from removable media](https://help.gnome.org/gnome-help/files-autorun.html). A deliberate launch works without installing a background service.

These comparisons describe the linked projects' documentation. They are not independent security audits.
