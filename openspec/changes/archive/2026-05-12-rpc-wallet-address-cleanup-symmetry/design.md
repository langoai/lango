## Overview

This is a small test-only symmetry improvement.

## Design Decisions

### Close the last obvious address lifecycle gaps

The suite already covered address success, sender error, and context cancel cleanup. The remaining obvious gaps were:

- companion error response
- timeout

Those are now covered directly so the address request lifecycle matches the rigor of the signing paths.
