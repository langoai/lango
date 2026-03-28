## ADDED Requirements

### Requirement: ChatParts struct for composable rendering
ChatModel SHALL expose a public RenderParts() method returning a ChatParts struct with Header, TurnStrip, Main, Footer, and Approval string fields.

#### Scenario: RenderParts returns all sections
- **WHEN** RenderParts() is called on a ChatModel with width > 0
- **THEN** it SHALL return ChatParts with non-empty Header, TurnStrip, Main, and Footer fields

#### Scenario: View uses RenderParts internally
- **WHEN** View() is called
- **THEN** it SHALL call RenderParts() and join the non-empty sections with newlines, producing identical output to the previous View() implementation
