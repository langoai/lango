## Overview

This is a small correctness alignment between existing workbench spec and key-routing behavior.

## Design Decisions

### Close the last focus-lane gap

Armed starter submission now treats `Decisions` the same way as the other empty-workbench lanes, preserving the “submit from any focus lane” user contract already carried by the main spec.
