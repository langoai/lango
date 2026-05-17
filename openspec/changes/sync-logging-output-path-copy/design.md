# Design: Sync Logging Output Path Copy

## Approach

Treat `logging.outputPath` copy as downstream UI/docs surface for the logging default writer behavior. Keep the implementation limited to settings form text, public configuration docs, and an executable form assertion.

## Test Strategy

Extend the existing `TestNewLoggingForm_AllFields` coverage to assert that the output-path placeholder and description mention stderr and no longer describe empty output as stdout.

## Compatibility

No config schema changes are introduced. Existing empty `logging.outputPath` values continue to use the runtime default logging writer.
