## Overview

This is a narrow regression-hardening pass.

## Design Decisions

### Close only the remaining address-path coverage gaps

The suite already covered address cleanup after success, sender error, and context cancellation. This change adds the remaining obvious address exit classes without broadening scope beyond what is already implemented.
