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
to pick the right one, or you can qualify a duplicate name by device — for
example `reviewer@blog`.

Replies come back the same way: the recipient can "tell" you back through the
same path, and the exchange stays threaded.

## Good to know

- **Discovery.** Sessions are addressed by their handle (custom name or project),
  so you refer to them the way you already think of them; UUIDs never come up.
- **Busy recipients.** If the target session is mid-task, the message waits and is
  delivered the moment it returns to its prompt — never dropped, never injected
  mid-turn.
- **Stay skeptical.** A delivered message is labelled as relayed from another
  session and treated as untrusted input, so the recipient uses judgment before
  acting. Every message is also kept in a durable log.
- **Scope today.** Claude Code and OpenCode sessions can both send and receive.
  Both sessions must report to the same dashboard.
