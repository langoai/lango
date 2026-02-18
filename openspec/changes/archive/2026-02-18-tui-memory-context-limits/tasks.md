## 1. Settings Form

- [x] 1.1 Add `om_max_reflections` field (InputInt, non-negative validation) to `NewObservationalMemoryForm` in `forms_impl.go`
- [x] 1.2 Add `om_max_observations` field (InputInt, non-negative validation) to `NewObservationalMemoryForm` in `forms_impl.go`

## 2. State Update Handler

- [x] 2.1 Add `om_max_reflections` case to `UpdateConfigFromForm` in `state_update.go`
- [x] 2.2 Add `om_max_observations` case to `UpdateConfigFromForm` in `state_update.go`
