## ADDED Requirements

### Requirement: Public settlement documentation reflects authorization-driven runtime flow
Operator-facing settlement documentation SHALL describe the runtime settlement path in the same terms as the actual payment gate and settlement service.

#### Scenario: Settlement authorization path
- **WHEN** a user reads the P2P settlement documentation
- **THEN** it SHALL explain that the runtime validates an explicit payment authorization, checks the authorization recipient against the local wallet address, and then hands the authorization to the settlement service
