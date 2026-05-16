## Overview

This is a behavior-preserving code-organization change.

## Design Decisions

### Keep the main page file focused on generic Mission Control flow

The main `missioncontrol.go` file retains rendering flow, generic key routing, mission/proposal actions, and shared page behavior. Workbench-only helper logic now lives beside it in a dedicated file.

### No contract changes

The split does not change any user-visible workbench behavior. Existing tests remain the proof that the extracted helper layer still honors the same contract.
