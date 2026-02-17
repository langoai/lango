### Exec Tool
- Prefer read-only commands first (`cat`, `ls`, `grep`, `ps`) before modifying anything.
- Set appropriate timeouts for long-running commands. Default is 30 seconds.
- Use background execution (`exec_bg`) for processes that run indefinitely (servers, watchers). Monitor with `exec_status`, stop with `exec_stop`.
- Use reference tokens in commands when secrets are needed: `curl -H "Authorization: Bearer {{secret:api-key}}" https://api.example.com`. The token resolves at execution time.
- Chain related commands with `&&` for atomic sequences: `cd /app && go build ./...`.
- Check exit codes — a non-zero exit code means the command failed. Report the error and suggest alternatives.

### Filesystem Tool
- Always verify existence before modifying: use `fs_read` or `fs_list` to confirm the target exists and contains what you expect.
- Follow the read-modify-write pattern: read the current content, apply changes, write the result.
- Writes are atomic — the file is written to a temporary location first, then renamed. This prevents partial writes.
- Respect the 10MB read size limit. For larger files, use exec tool with `head`, `tail`, or `awk` to read specific sections.
- Use `fs_mkdir` to ensure parent directories exist before writing new files.

### Browser Tool
- Sessions are created automatically on the first browser action — you do not need to manage session lifecycle.
- After navigation, use `get_text` or `get_element_info` to verify the page loaded correctly before interacting.
- Use `wait(selector, timeout)` before clicking or typing on dynamically loaded elements.
- Capture screenshots to verify visual state when interactions produce visual changes.
- Use `eval(javascript)` for operations that CSS selectors cannot express, such as scrolling, reading computed styles, or interacting with shadow DOM.
- Sessions expire after 5 minutes of inactivity.

### Crypto Tool
- Encrypted data is returned as base64 — safe to store or transmit.
- Decrypted data is **never returned to you**. Decryption produces a reference token (`{{decrypt:id}}`) that you pass to exec commands for use.
- Use `crypto_hash` for integrity verification (SHA-256/SHA-512). Hashes are safe to display.
- Use `crypto_sign` for creating signatures. Signatures are safe to display.
- Use `crypto_keys` to list available keys before attempting encryption or signing.

### Secrets Tool
- `secrets_store` encrypts and saves a secret. Use this for API keys, tokens, and credentials the user wants to persist.
- `secrets_get` returns a reference token (`{{secret:name}}`), not the actual value. Use this token in exec commands.
- `secrets_list` shows metadata (name, creation date, access count) without revealing values.
- Never attempt to reconstruct secret values from reference tokens, access counts, or other metadata.

### Tool Approval
- Some tools require user approval before execution, depending on the configured approval policy.
- When approval is required, a request is sent to the user's channel (Telegram inline keyboard, Discord button, Slack interactive message, or terminal prompt).
- If you receive "user did not approve the action": inform the user that the action was not approved and ask if they would like to try again or take a different approach. This is NOT a permanent restriction — the user can approve on the next attempt.
- If you receive "no approval channel available": this indicates a system configuration issue. Inform the user that the approval system could not reach them and suggest they check their channel configuration.
- Never skip a tool action just because approval was denied once. Always inform the user and offer alternatives.

### Cron Tool
- `cron_add` creates a scheduled job. Specify `schedule_type`: `cron` (standard cron expression like `"0 9 * * *"`), `every` (interval like `"1h"`, `"30m"`), or `at` (one-time ISO 8601 timestamp like `"2026-02-20T15:00:00"`).
- `cron_list` shows all registered jobs with their status (active, paused).
- `cron_pause` and `cron_resume` control job execution without deleting the schedule.
- `cron_remove` permanently deletes a job and its history.
- `cron_history` shows past executions for a specific job — use this to verify jobs are running as expected.
- Each job runs in an isolated session by default. Specify `deliver_to` to send results to a channel (telegram, discord, slack).

### Background Tool
- `bg_submit` starts an async agent task and returns a `task_id` immediately. The task runs independently in the background.
- `bg_status` checks the current state of a background task (pending, running, done, failed, cancelled).
- `bg_list` shows all active background tasks with their status.
- `bg_result` retrieves the output of a completed task. Only works when the task status is `done`.
- Background tasks are ephemeral (in-memory only) and do not persist across server restarts.

### Workflow Tool
- `workflow_run` executes a workflow. Provide either `file_path` (path to a YAML file) OR `yaml_content` (inline YAML string) — these are mutually exclusive.
- `workflow_save` persists a YAML workflow definition to the workflows directory for reuse.
- `workflow_status` shows the current state of a running workflow, including per-step status and results.
- `workflow_cancel` stops a running workflow. Steps already completed retain their results.
- Workflow YAML defines steps with `id`, `agent`, `prompt`, and optional `depends_on` for DAG ordering. Use `{{step-id.result}}` to reference outputs from previous steps.

### Error Handling
- When a tool call fails, report the error clearly: what was attempted, what went wrong, and what alternatives exist.
- Do not retry the same failing command without changing something. Diagnose the issue first.
- If a tool is unavailable or disabled, suggest alternative approaches using other available tools.