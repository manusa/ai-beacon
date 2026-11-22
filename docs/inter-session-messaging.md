# Inter-session messaging

Stop being the courier between terminals. When one session opens a PR and another
reviews it, you normally copy-paste the review back yourself. AI Beacon lets the
sessions talk directly: from inside session B you say, in plain language, "tell
Session A about points 1 and 2", and session A receives it as a normal instruction
and just acts. You stay the router and approver, not the messenger.

## Using it

Nothing to install or configure — any session launched with `ai-beacon session
...` gets the messaging tools automatically.

To send, just tell your agent in plain language:

> tell the reviewer session about action points 1 and 2, skip the rest

Your agent resolves who "the reviewer session" is, distills what you asked for,
and sends it. The recipient session receives it as a natural prompt — framed so
it knows the message came from another session, relayed by you — and continues
working on it. No command needed on the receiving side.

If the target is vague or ambiguous, your agent can list the reachable sessions
to pick the right one, or you can qualify a duplicate name by project or device
— for example `reviewer@blog`. When several sessions share one repo on one machine,
your agent asks you which one you mean rather than guessing.

Replies come back the same way: the recipient can "tell" you back through the
same path, and the exchange stays threaded.

## Good to know

- **Discovery.** You address sessions by their handle: the custom name, else the
  name you gave the agent (Claude Code's `/rename`), else the project, so you refer
  to them the way you already think of them. A device's
  [Oracle](multi-machine.md#an-oracle-per-device) also answers to
  `oracle@<device>`, where the device can be typed in a short form: lowercase,
  letters and digits only, with dashes in between, so the Oracle on
  `🐧 nuc-001 (WSL)` is `oracle@nuc-001-wsl`. Your agent sees that short form in
  its session list. A bare `oracle` reaches your own system's Oracle when it has
  exactly one, even with Oracles on other systems. Agents use each session's
  id under the hood (to reply, and to tell apart two sessions on one repo), but you
  never have to type one.
- **Busy recipients.** If the target session is mid-task, the message waits and is
  delivered the moment it returns to its prompt — never dropped, never injected
  mid-turn.
- **Stay skeptical.** A delivered message is labelled as relayed from another
  session and treated as untrusted input, so the recipient uses judgment before
  acting. Every message is also kept in a durable log.
- **Scope today.** Claude Code and OpenCode sessions can both send and receive.
  Both sessions must report to the same dashboard. OpenCode 2.0 and 2.1 offer the
  messaging tools only inside their Code Mode (the model calls them from code it
  runs in the `execute` tool); OpenCode 2.2 and later, like OpenCode 1, offer them
  as ordinary tools. See
  [opencode 2](connect-agent.md#opencode-2) for details.
