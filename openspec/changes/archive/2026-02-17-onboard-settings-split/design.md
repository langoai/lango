# Design: Onboard/Settings Split

## Architecture

```
internal/cli/
  tuicore/             # Shared form components
    field.go           # Field, InputType types (TextInput exported)
    form.go            # FormModel (Init/Update/View)
    state.go           # ConfigState + dirty tracking
    state_update.go    # UpdateConfigFromForm/UpdateProviderFromForm/UpdateAuthProviderFromForm

  onboard/             # 5-step wizard
    onboard.go         # Cobra command
    wizard.go          # 5-step stepper model
    progress.go        # Progress bar + step list renderer
    steps.go           # Wizard step form constructors (essential fields only)
    test_step.go       # Configuration validation step

  settings/            # Full menu-based editor
    settings.go        # Cobra command (lango settings --profile)
    editor.go          # Editor model (renamed from Wizard)
    menu.go            # Configuration menu (17 categories)
    forms_impl.go      # All form constructors
    providers_list.go  # Provider management UI
    auth_providers_list.go  # OIDC provider management UI
```

## Key Decisions

### 1. Shared tuicore package
Extracted `Field`, `FormModel`, `ConfigState`, and state update functions into a shared package to avoid code duplication. The `textInput` field was exported as `TextInput` for cross-package access.

### 2. Wizard Steps
The onboard wizard has 5 steps:
1. **Provider Setup** — Provider type, name, API key, base URL
2. **Agent Config** — Provider selection, model, max tokens, temperature
3. **Channel Setup** — Channel selector (Telegram/Discord/Slack/Skip) → channel-specific form
4. **Security & Auth** — Privacy interceptor, PII redaction, approval policy
5. **Test Configuration** — Validates provider exists, API key set, model set, channel tokens, config.Validate()

### 3. Navigation
- `Ctrl+N`: Save current form → advance to next step
- `Ctrl+P`: Save current form → go back one step
- `Ctrl+C`: Cancel and quit
- `Esc`: Context-dependent (quit on Step 1, go back on others)

### 4. Channel Selection
Step 3 uses a custom channel selector (not a FormModel) before showing the channel-specific form. "Skip" advances to Step 4 without configuring any channel.

### 5. Test Results
Step 5 auto-runs 5 validation checks using `tui.FormatPass/FormatWarn/FormatFail` indicators. Press Enter to save and complete.
