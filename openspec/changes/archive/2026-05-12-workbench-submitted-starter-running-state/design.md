## Overview

This is a UX-state alignment change for the moment immediately after a starter prompt is submitted.

## Design Decisions

### Running-state guidance wins over quick-start guidance

When the workbench is still otherwise empty but the shared chat model reports an in-flight turn, the empty-state body, placeholder, and footer now switch to a running-state hint. This prevents the UI from advertising startup actions while the user is already waiting on a result.
