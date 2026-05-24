## Design Summary

This slice extends the cockpit dead-letter page with unified selection preservation across:

- `Enter` apply
- `Ctrl+R` reset
- retry-success refresh

Rule:

- if the current `selectedID` still exists in the refreshed result set, preserve it

Fallback:

- first row when results exist
- clear selection/detail when the refreshed result set is empty
