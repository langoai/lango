## Why

The quickstart guide still tells users that bare `lango` launches the cockpit. That has been false since the workbench split: bare `lango` launches the standalone mission workbench, while `lango cockpit` launches the explicit multi-panel dashboard.

This is a high-visibility onboarding doc, so the wording needs to match the actual entrypoint behavior.

## What Changes

- Update the quickstart tip so bare `lango` is described as the mission workbench.
- Mention `lango cockpit` as the explicit multi-panel dashboard path.
- Extend downstream docs requirements so the quickstart guide stays aligned with the workbench/cockpit split.

## Impact

- New users get the correct first-run expectation from the quickstart guide.
- Public onboarding docs match the current CLI surface.
