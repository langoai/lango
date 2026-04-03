# Catastrophic Pattern Expansion — Tasks

## Implementation

- [x] 1.1 Add privilege escalation patterns (sudo, su, chmod +s, chown root)
- [x] 1.2 Add remote code execution patterns (curl|sh, wget|bash)
- [x] 1.3 Add reverse shell patterns (nc -l, ncat, socat)
- [x] 1.4 Add block device and mass deletion patterns
- [x] 1.5 Add `ObservePatterns` field and `DefaultObservePatterns()`
- [x] 1.6 Add compound pattern pre-computation at construction time
- [x] 1.7 Extract shared `matchPattern()` helper
- [x] 1.8 Wire observe patterns through hook middleware

## Testing

- [x] 2.1 Test all new blocked patterns match correctly
- [x] 2.2 Test observe patterns log but don't block
- [x] 2.3 Test compound patterns (multi-part matching)
- [x] 2.4 Test backward compatibility (existing patterns still work)
