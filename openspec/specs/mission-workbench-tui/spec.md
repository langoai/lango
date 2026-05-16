# mission-workbench-tui Specification

## Purpose
Define the standalone `lango` mission workbench contract, including its lighter Mission Control surface and first-run empty-state guidance.
## Requirements
### Requirement: Bare `lango` launches the mission workbench

Running `lango` without a subcommand on an interactive terminal SHALL launch a standalone mission workbench rather than the cockpit shell or the focused chat surface.

#### Scenario: Bare `lango` launches the workbench
- **WHEN** the user runs `lango` on an interactive terminal
- **THEN** the application SHALL open the mission workbench surface
- **AND** that surface SHALL become the bare-`lango` contract for Wave 6

### Requirement: The mission workbench hosts Mission Control content without cockpit chrome

The mission workbench SHALL present Mission Control content directly without the full cockpit sidebar or context-panel shell. It reuses Mission Control behavior, but it is not itself the cockpit shell.

#### Scenario: Workbench shows Mission Control content directly
- **WHEN** the workbench renders successfully
- **THEN** the user SHALL see Mission Control content such as missions, live decision state, loops, activity, and the shared composer
- **AND** the workbench SHALL NOT require the full cockpit sidebar or context-panel chrome to expose that content

### Requirement: The workbench remains a lighter local surface than cockpit

The first Wave 6 workbench slice SHALL stay local and mission-native. It may reuse the same Mission Control projection assets as cockpit, but it SHALL NOT imply that cockpit-only surfaces or channel startup belong to bare `lango`.

#### Scenario: Workbench hints to the other explicit surfaces
- **WHEN** the workbench renders first-screen copy or help
- **THEN** it SHALL hint to `lango chat` as the focused chat surface
- **AND** it SHALL hint to `lango cockpit` as the advanced dashboard

#### Scenario: Cockpit-only channel startup is not implied by bare `lango`
- **WHEN** the user launches bare `lango`
- **THEN** the first Wave 6 slice SHALL NOT imply that `--with-channels` or live channel startup belongs to the workbench contract

### Requirement: Workbench empty state guides incomplete profiles
The standalone workbench empty state SHALL guide the operator toward setup and verification when the active profile is obviously incomplete.

#### Scenario: Incomplete profile shows setup guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile
- **THEN** the empty state SHALL mention `lango onboard`, `lango settings`, and `lango doctor`
- **AND** SHALL keep the existing chat and cockpit navigation hints

### Requirement: Workbench empty state stays clean for ready profiles
The standalone workbench empty state SHALL omit setup guidance when the active profile already has a usable provider/model path.

#### Scenario: Ready profile omits setup guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty state SHALL NOT mention `lango onboard`, `lango settings`, or `lango doctor`

### Requirement: Ready-profile workbench empty state suggests starter prompts
The standalone workbench empty state SHALL suggest concrete starter prompts when the active profile is ready for normal use.

#### Scenario: Ready profile shows starter prompts
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty state SHALL mention starter prompts for repository summary, project-structure explanation, and recent-change review
- **AND** SHALL keep the existing chat and cockpit navigation hints

### Requirement: Workbench composer placeholder follows profile readiness
The standalone workbench composer placeholder SHALL mirror the same readiness split as the workbench empty-state body when the composer is empty.

#### Scenario: Incomplete profile shows setup-first composer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile and the composer is empty
- **THEN** the composer placeholder SHALL instruct the operator to use `lango onboard`, `lango settings`, or `lango doctor`

#### Scenario: Ready profile shows starter-prompt composer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and the composer is empty
- **THEN** the composer placeholder SHALL suggest starter prompts for repository summary, project-structure explanation, and recent-change review

### Requirement: Workbench header reflects incomplete profile setup
The standalone workbench header SHALL show a setup-required model summary when the active profile does not yet have a usable provider/model path.

#### Scenario: Incomplete profile shows setup-required header
- **WHEN** bare `lango` renders the workbench with an incomplete profile
- **THEN** the header SHALL show `Model: Setup required`
- **AND** the empty-state body and composer guidance SHALL continue pointing the operator to setup recovery commands

### Requirement: Workbench header preserves ready profile summary
The standalone workbench header SHALL keep the concrete provider/model summary for ready profiles.

#### Scenario: Ready profile keeps provider and model summary
- **WHEN** bare `lango` renders the workbench with a ready profile
- **THEN** the header SHALL show the configured provider and model summary instead of `Setup required`

### Requirement: Workbench readiness signals share one profile-readiness contract

The standalone workbench SHALL derive its setup-required header, empty-state guidance, and composer placeholder from one shared agent-readiness evaluation.

#### Scenario: Missing remote API key keeps all workbench recovery cues aligned
- **WHEN** bare `lango` renders with a non-ollama provider ID and model configured but the provider API key is empty
- **THEN** the header SHALL show `Model: Setup required`
- **AND** the empty-state body SHALL continue setup guidance
- **AND** the empty composer placeholder SHALL continue setup-first guidance

#### Scenario: Ollama remains ready without an API key
- **WHEN** bare `lango` renders with an ollama provider ID and model configured but no API key
- **THEN** the header SHALL show the configured provider and model summary
- **AND** the empty-state body SHALL omit setup guidance
- **AND** the empty composer placeholder SHALL show starter prompts

### Requirement: Ready-profile workbench starter prompts are keyboard-addressable

The standalone workbench SHALL let the operator load the ready-profile starter prompts through direct keyboard shortcuts instead of forcing retyping from the empty-state copy.

#### Scenario: Starter hotkeys load prompts into the empty composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and an empty composer
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the corresponding starter prompt SHALL be loaded into the composer
- **AND** the operator SHALL remain in control of whether to press `Enter` to run it

#### Scenario: Starter hotkeys are advertised in ready-profile workbench copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and an empty composer
- **THEN** the empty-state body SHALL mention the `1`, `2`, and `3` starter-prompt hotkeys
- **AND** the empty composer placeholder SHALL mention the `Press 1-3` quick-start path
- **AND** the footer SHALL expose the starter-prompt hotkeys while that state is active

### Requirement: Ready-profile workbench starter prompts reflect workspace context

The standalone workbench SHALL adapt its ready-profile starter prompts to the detected workspace context when `lango` starts inside a project.

#### Scenario: Repository-aware prompts in a detected project
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workdir belongs to a repository
- **THEN** the starter prompts SHALL reference the detected repository instead of using only generic copy

#### Scenario: Go-aware structure prompt in a Go module
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workdir belongs to a workspace containing a `go.mod`
- **THEN** the structure-oriented starter prompt SHALL use Go package layout guidance instead of a generic project-structure prompt

#### Scenario: Generic fallback outside a detected project
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** no repository markers are detected from the current workdir
- **THEN** the starter prompts SHALL fall back to the generic repository-summary, project-structure, and recent-change prompts

### Requirement: Ready-profile workbench starter prompts reflect live Git state when available

The standalone workbench SHALL use lightweight Git signals to sharpen the ready-profile change-review starter prompt when the detected workspace is a Git repository and Git metadata is available.

#### Scenario: Clean repository prompt mentions the current branch
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with a current branch and no uncommitted changes
- **THEN** the change-review starter prompt SHALL mention the current branch

#### Scenario: Dirty repository prompt mentions uncommitted changes
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **THEN** the change-review starter prompt SHALL mention the uncommitted changes and the current branch

#### Scenario: Git failure keeps the repository-aware fallback
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is repository-shaped but Git metadata cannot be read
- **THEN** the change-review starter prompt SHALL fall back to the repository-aware non-Git wording instead of failing the workbench startup path

### Requirement: Dirty-repository workbench starter prompts mention changed targets when available

The standalone workbench SHALL mention the most obvious changed files or directories in the dirty-repository starter prompt when lightweight Git status output can be summarized.

#### Scenario: Dirty repository prompt highlights changed targets
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **AND** the changed targets can be summarized from lightweight Git status output
- **THEN** the dirty-repository starter prompt SHALL mention the current branch
- **AND** SHALL mention the summarized changed files or directories

#### Scenario: Dirty repository prompt falls back when changed targets are unclear
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **AND** changed targets cannot be summarized
- **THEN** the dirty-repository starter prompt SHALL still mention the current branch and uncommitted changes
- **AND** SHALL NOT fail the startup path

### Requirement: Workbench starter prompts share one generation contract

The standalone workbench SHALL derive starter prompt defaults and starter prompt context-shaping from one shared generation contract rather than from duplicated page-local fallback strings.

#### Scenario: Shared default prompts backstop the workbench page
- **WHEN** the Mission Control workbench page needs its fallback starter prompts
- **THEN** it SHALL use the same shared starter prompt contract that the workbench shell uses to build context-aware prompt sets
- **AND** quick-start copy SHALL remain behaviorally consistent across the workbench shell and page rendering path

### Requirement: Workbench prompt derivation depends on stable workspace inputs

The standalone workbench SHALL derive starter prompt behavior from stable workspace inputs such as `workDir` through the shared prompt helper, rather than relying on a separately transported precomputed prompt slice.

#### Scenario: Mission Control page derives prompts from workDir
- **WHEN** the workbench Mission Control page needs ready-profile starter prompts
- **THEN** it SHALL derive them from the shared prompt helper using the current workspace input
- **AND** the starter prompt behavior SHALL stay consistent with the rest of the workbench shell

### Requirement: Ready-profile workbench Enter key seeds the default starter prompt

The standalone workbench SHALL let the operator use `Enter` as the default quick-start key on an empty ready-profile first screen.

#### Scenario: Enter seeds the first starter prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the composer is empty
- **AND** the operator presses `Enter`
- **THEN** the first starter prompt SHALL be loaded into the composer
- **AND** the operator SHALL remain in control of whether to press `Enter` again to submit it

#### Scenario: Incomplete profile does not seed a starter prompt on Enter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL keep the setup-first guidance instead of seeding a starter prompt

### Requirement: Ready-profile workbench copy advertises the Enter quick-start path

The standalone workbench SHALL explicitly advertise `Enter` as the default quick-start seed on the empty ready-profile first screen.

#### Scenario: Ready-profile copy mentions Enter and numeric hotkeys
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty-state body SHALL mention `Enter` for the default starter prompt
- **AND** SHALL continue mentioning `1`, `2`, and `3` for explicit starter selection
- **AND** the footer or empty composer hint SHALL surface the same `Enter` quick-start path

### Requirement: Enter quick-start default follows repository state

The standalone workbench SHALL choose the default `Enter` quick-start prompt from workspace context instead of always using the summary prompt.

#### Scenario: Dirty repository defaults Enter to change review
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a dirty Git repository
- **AND** the operator presses `Enter`
- **THEN** the seeded default prompt SHALL be the context-aware change-review prompt

#### Scenario: Clean workspace defaults Enter to summary
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is clean or not repository-backed
- **AND** the operator presses `Enter`
- **THEN** the seeded default prompt SHALL remain the summary-oriented default prompt

### Requirement: Seeded starter prompts advertise the submit step

The standalone workbench SHALL distinguish between the starter-seeding step and the starter-submission step in its quick-start copy.

#### Scenario: Seeded starter shows submit-focused footer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has already been seeded into the composer
- **THEN** the footer hint SHALL indicate that `Enter` submits the starter prompt
- **AND** it SHALL stop showing the pre-seed `Enter default starter` hint for that state

### Requirement: Seeded starter prompts switch body copy to submit guidance

The standalone workbench SHALL replace the ready-profile quick-start body copy with submit guidance once a starter prompt is armed in the composer.

#### Scenario: Seeded starter body shows submit guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has been seeded into the composer
- **THEN** the empty-state body SHALL explain that `Enter` runs the starter prompt or that the operator may edit it before sending
- **AND** it SHALL stop showing the pre-seed quick-start line for that state

### Requirement: Workbench starter prompt derivation is cached per page instance

The standalone workbench SHALL derive its starter prompt contract from workspace inputs once per Mission Control page instance rather than repeatedly recomputing it during ordinary render work.

#### Scenario: Cached starter prompt contract backs render-time copy
- **WHEN** a Mission Control workbench page is created for a given workspace input
- **THEN** the page SHALL cache the derived starter prompt set and default starter prompt for that page instance
- **AND** ordinary render-time quick-start copy SHALL reuse that cached contract

### Requirement: Seeded starter guidance reflects whether Composer still has focus

The standalone workbench SHALL tailor seeded-starter guidance to the current focus lane so it only tells the operator to press keys that will actually work next.

#### Scenario: Seeded starter outside composer explains the focus step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus is no longer on the composer lane
- **THEN** the guidance SHALL instruct the operator to return to `Composer` before pressing `Enter`

### Requirement: Seeded starter prompts submit from any empty-workbench focus lane

The standalone workbench SHALL let the operator submit an armed starter prompt with `Enter` even if focus has moved away from `Composer`, as long as the workbench is still in the empty ready-profile seeded state.

#### Scenario: Seeded starter submits outside composer focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus has moved to a different lane
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL submit the armed starter prompt
- **AND** seeded-state guidance SHALL no longer require tabbing back to `Composer` first

### Requirement: Armed starter prompts remain replaceable by numeric starter shortcuts

The standalone workbench SHALL let the operator replace an already armed starter prompt with `1`, `2`, or `3` instead of treating those keys as plain text input.

#### Scenario: Numeric starter shortcut replaces an armed starter prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already armed in the composer
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the composer SHALL switch to the corresponding starter prompt
- **AND** the keypress SHALL NOT append the digit as free text

### Requirement: Mission Control isolates workbench-only quick-start helpers from generic page flow

The workbench-specific quick-start and setup-recovery helper layer SHALL remain isolated from the generic Mission Control page flow so future workbench UX changes do not further crowd the shared page source.

#### Scenario: Workbench helpers live in a dedicated companion source
- **WHEN** the Mission Control page source is organized for maintenance
- **THEN** workbench-only starter/setup helper logic SHALL live in a dedicated companion source file rather than continuing to accumulate inside the primary Mission Control page source

### Requirement: Armed starter submission works from Decisions focus too

The standalone workbench SHALL honor the seeded-starter any-focus submit contract specifically from the `Decisions` lane as well as the other empty-workbench lanes.

#### Scenario: Armed starter submits from Decisions focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already armed in the composer
- **AND** focus is on `Decisions`
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL submit the armed starter prompt

### Requirement: Armed starter editing keys return focus to Composer

The standalone workbench SHALL treat composer editing keys as intent to edit the armed starter prompt even if focus has moved away from `Composer`.

#### Scenario: Armed starter backspace returns focus to Composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus has moved away from `Composer`
- **AND** the operator presses a composer editing key such as `Backspace`
- **THEN** the workbench SHALL return focus to `Composer`
- **AND** SHALL apply the edit to the armed starter prompt

### Requirement: Submitted starter prompts switch the empty workbench into a running-state hint

The standalone workbench SHALL replace its starter-oriented empty-state guidance with a running-state hint while a submitted starter prompt is still in flight.

#### Scenario: Submitted starter shows running-state guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has already been submitted
- **AND** the turn is still in flight
- **THEN** the empty-state body, footer, or placeholder SHALL indicate that the current request is running
- **AND** the workbench SHALL stop showing pre-submit quick-start guidance for that state

### Requirement: Running-state workbench copy advertises next-prompt staging

The standalone workbench SHALL describe the next-prompt staging path while a submitted starter prompt is still in flight.

#### Scenario: Running-state copy mentions interrupt-and-run follow-up
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a submitted starter prompt is still streaming
- **THEN** the running-state guidance SHALL mention typing the next prompt
- **AND** SHALL mention pressing `Enter` to interrupt and run it

### Requirement: Running-state Enter can queue the default follow-up when no draft exists

The standalone workbench SHALL let the operator use `Enter` as the zero-input follow-up path while a starter turn is still running.

#### Scenario: Running-state Enter queues the default follow-up
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** no follow-up draft has been staged yet
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL queue and run the default context-aware follow-up prompt as the next turn

### Requirement: Running-state follow-up guidance preserves direct editing

The standalone workbench SHALL treat staged follow-up editing as part of the running-state interaction loop, not as a separate hidden behavior.

#### Scenario: Running-state follow-up editing returns to Composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator presses an editing key while focus is away from `Composer`
- **THEN** the workbench SHALL return focus to `Composer`
- **AND** SHALL apply the edit to the staged follow-up prompt

### Requirement: Running-state follow-up guidance preserves starter replacement

The standalone workbench SHALL treat starter replacement as part of the running follow-up loop while a starter turn is still in flight.

#### Scenario: Running follow-up accepts starter replacement
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the staged follow-up SHALL switch to the corresponding starter prompt

### Requirement: Running follow-up drafts remain replaceable by starter hotkeys

The standalone workbench SHALL let the operator replace a staged follow-up draft with `1`, `2`, or `3` while the current starter turn is still running.

#### Scenario: Running follow-up hotkey replacement
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up draft
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the staged follow-up SHALL switch to the corresponding starter prompt
- **AND** the digit SHALL NOT be appended as free text

### Requirement: Running follow-up starter replacement changes the next turn that runs

The standalone workbench SHALL treat starter replacement during a running follow-up loop as a change to the actual next turn that will execute.

#### Scenario: Replaced follow-up starter becomes the next executed prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator replaces that follow-up with `1`, `2`, or `3`
- **AND** the operator presses `Enter`
- **THEN** the replacement starter prompt SHALL be the next turn that runs

### Requirement: Running follow-up drafts submit from any empty-workbench focus lane

The standalone workbench SHALL let the operator submit a staged follow-up draft with `Enter` even if focus has moved away from `Composer`, as long as the workbench is still in the empty ready-profile running state.

#### Scenario: Staged follow-up submits from Decisions focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up draft
- **AND** focus is on `Decisions`
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL queue and run the staged follow-up as the next turn

### Requirement: Workbench activity retains assistant reply summaries after a turn completes

The standalone workbench SHALL retain a concise assistant reply summary in its activity lane after a turn completes so the operator can still see what happened from the workbench shell.

#### Scenario: Activity projection preserves compact reply summaries
- **WHEN** a compacted assistant activity summary is projected into the Mission Control activity view
- **THEN** the projected activity row SHALL preserve the one-line bounded summary instead of re-expanding the raw reply text

### Requirement: Post-turn empty workbench defaults Enter to the next-step starter

The standalone workbench SHALL shift its default empty-state `Enter` starter after at least one turn completes so the next loop starts from a next-step prompt instead of repeating the initial repository summary prompt.

#### Scenario: Completed turn changes the empty-state default starter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** there is no staged follow-up draft and no armed starter prompt
- **THEN** the empty-state copy SHALL advertise the next-step starter as the default `Enter` prompt

#### Scenario: Enter seeds the next-step starter after a completed turn
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** there is no staged follow-up draft and no armed starter prompt
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL arm the next-step starter prompt instead of the original repository-summary starter

### Requirement: Post-turn empty workbench default stays context-appropriate

The standalone workbench SHALL refine its post-turn default `Enter` starter so it still feels like the next step in the current workspace instead of falling back to a mismatched generic review prompt.

#### Scenario: Generic workspace uses structure-oriented post-turn default
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** no detected repo or workspace context is available
- **THEN** the empty-state default `Enter` starter SHALL pivot to the structure-oriented starter instead of the generic recent-changes starter

#### Scenario: Detected workspace keeps next-change post-turn default
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** detected repo or workspace context is available
- **THEN** the empty-state default `Enter` starter SHALL remain the next-change starter

### Requirement: Post-turn empty workbench copy names the next-step loop explicitly

The standalone workbench SHALL describe the completed-turn empty state as a next-step interaction loop rather than reusing first-run quick-start wording.

#### Scenario: Completed turn uses next-step copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body or footer SHALL describe the default `Enter` prompt as the next step instead of the initial quick start

### Requirement: Empty standalone workbench suppresses degraded-note noise

The standalone workbench SHALL avoid surfacing cockpit-style degraded warnings on its empty first screen when no active mission/control content exists yet.

#### Scenario: Empty workbench hides degraded note
- **WHEN** bare `lango` renders an empty Mission Control workbench state
- **AND** the projected header contains a degraded note
- **THEN** the empty workbench SHALL hide that degraded note from the first-screen shell

#### Scenario: Cockpit surface still shows degraded note
- **WHEN** the explicit `lango cockpit` Mission Control page renders an empty state with a degraded note
- **THEN** the degraded note SHALL remain visible

### Requirement: Completed-turn empty workbench body names the next-step state

The standalone workbench SHALL describe the completed-turn empty body as a next-step loop instead of the generic no-missions empty state.

#### Scenario: Completed-turn body names the next step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body SHALL indicate that the last turn completed and the next step is ready

### Requirement: Completed-turn empty workbench hint invites the next prompt explicitly

The standalone workbench SHALL update its completed-turn empty-state hint so it explicitly invites the next prompt instead of generic chat wording.

#### Scenario: Completed-turn hint says next prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state hint SHALL tell the operator to type the next prompt here

### Requirement: Completed-turn empty composer placeholder uses next-step wording

The standalone workbench SHALL keep its completed-turn composer placeholder aligned with the next-step body and footer wording.

#### Scenario: Completed-turn placeholder says next step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty composer placeholder SHALL say `Next step: press Enter ...` instead of the original first-run wording

### Requirement: Completed-turn workbench footer uses next-prompt wording

The standalone workbench SHALL keep its completed-turn footer aligned with the next-step body and placeholder wording.

#### Scenario: Completed-turn footer says next prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the footer SHALL tell the operator to type the next prompt here instead of generic chat wording

### Requirement: Completed-turn empty workbench previews the last result

The standalone workbench SHALL surface the latest assistant summary directly in the completed-turn empty body so the operator can see what just happened without scanning the activity lane.

#### Scenario: Completed-turn body shows last result preview
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** a compact assistant activity summary is available
- **THEN** the empty body SHALL include that latest result as a compact last-result preview

### Requirement: Completed-turn empty body presents the primary next step before the typing hint

The standalone workbench SHALL present the completed-turn next-step starter guidance before the generic next-prompt typing hint so the recommended action reads first.

#### Scenario: Completed-turn body orders starter before typing hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the completed-turn next-step starter guidance SHALL appear before the generic next-prompt hint in the empty body

### Requirement: Completed-turn empty body uses neutral finished wording

The standalone workbench SHALL describe the completed-turn empty state with neutral `finished` wording rather than success-specific `complete` wording.

#### Scenario: Completed-turn body says finished
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body SHALL say that the last turn finished rather than that it completed

### Requirement: Completed-turn empty body calls out failed turns explicitly

The standalone workbench SHALL distinguish a failed prior turn from a successful one in its completed-turn empty body lead.

#### Scenario: Failed turn changes the completed-turn lead
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the empty-state lead SHALL indicate that the last turn needs attention instead of using the neutral finished wording

### Requirement: Completed-turn result preview avoids redundant assistant labels

The standalone workbench SHALL avoid repeating the assistant label inside the completed-turn body preview when the preview already has a `Last result:` prefix.

#### Scenario: Success preview drops assistant label duplication
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest completed-turn summary is an assistant success summary
- **THEN** the body SHALL show `Last result: <summary>` without repeating `Assistant reply:`

### Requirement: Failed completed-turn workbench copy uses recovery wording

The standalone workbench SHALL switch its completed-turn copy from generic next-step wording to recovery-specific wording when the latest turn failed.

#### Scenario: Failed turn uses recovery starter copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the body or placeholder SHALL describe the default `Enter` starter as a recovery step
- **AND** the footer SHALL describe `Enter` as a recovery starter

### Requirement: Failed completed-turn footer uses recovery-prompt wording

The standalone workbench SHALL keep its failed completed-turn footer aligned with the recovery-specific body and placeholder wording.

#### Scenario: Failed turn footer says recovery prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the footer SHALL say `Type recovery prompt here` instead of `Type next prompt here`

### Requirement: Failed completed-turn body lead uses recovery-step wording

The standalone workbench SHALL keep the failed completed-turn body lead aligned with the rest of the recovery-specific copy.

#### Scenario: Failed turn lead says recovery step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the empty-state lead SHALL tell the operator to pick the recovery step instead of the generic next step

### Requirement: Failed completed-turn Enter default uses a recovery-oriented starter

The standalone workbench SHALL seed a recovery-oriented default starter when the latest completed turn failed and the operator presses `Enter` from the empty composer.

#### Scenario: Failed turn Enter default uses the recovery starter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **AND** the composer is empty
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL seed the recovery-oriented starter instead of reusing the generic completed-turn default

