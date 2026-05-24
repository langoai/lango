## Design Summary

This slice extends the cockpit dead-letter page with:

- a `Ctrl+R` global reset shortcut

Reset behavior:

- resets all draft filter state
- resets all applied filter state
- clears retry confirm state
- resets the active field to the default query field
- reloads backlog and selected detail using the existing first-row reset path

Guard behavior:

- if retry is `running`, `Ctrl+R` is ignored
