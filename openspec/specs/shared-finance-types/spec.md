## Purpose

Capability spec for shared-finance-types. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: finance package provides USDC parsing with shopspring/decimal
The `internal/finance` package SHALL provide `ParseUSDC(amount string) (*big.Int, error)` that converts a decimal string to the smallest USDC unit (6 decimals) using `shopspring/decimal` for exact arithmetic.

#### Scenario: Valid decimal string is parsed
- **WHEN** `ParseUSDC("1.50")` is called
- **THEN** it returns `1500000` as `*big.Int` with no error

#### Scenario: String with too many decimal places is rejected
- **WHEN** `ParseUSDC("0.1234567")` is called
- **THEN** it returns an error indicating too many decimal places

#### Scenario: Invalid string is rejected
- **WHEN** `ParseUSDC("abc")` is called
- **THEN** it returns an error indicating invalid USDC amount

### Requirement: finance package provides USDC formatting
The `internal/finance` package SHALL provide `FormatUSDC(amount *big.Int) string` that converts smallest USDC units back to a decimal string with trailing zero trimming (minimum 2 decimal places).

#### Scenario: Whole amount is formatted
- **WHEN** `FormatUSDC(big.NewInt(1_000_000))` is called
- **THEN** it returns `"1.00"`

#### Scenario: Fractional amount trims trailing zeros
- **WHEN** `FormatUSDC(big.NewInt(500_000))` is called
- **THEN** it returns `"0.50"`

#### Scenario: Sub-cent amount preserves precision
- **WHEN** `FormatUSDC(big.NewInt(1))` is called
- **THEN** it returns `"0.000001"`

### Requirement: finance package provides float-to-USDC conversion
The `internal/finance` package SHALL provide `FloatToMicroUSDC(amount float64) *big.Int` that converts a float64 dollar amount using `shopspring/decimal` for precision. The function MUST NOT use the `int64(amount * 1_000_000)` pattern.

#### Scenario: Standard float conversion
- **WHEN** `FloatToMicroUSDC(1.5)` is called
- **THEN** it returns `1500000` as `*big.Int`

#### Scenario: Edge-case float representation
- **WHEN** `FloatToMicroUSDC(0.000001)` is called
- **THEN** it returns `1` as `*big.Int` (not 0, which the truncation pattern would produce)

### Requirement: finance package provides shared constants
The `internal/finance` package SHALL export `USDCDecimals = 6`, `CurrencyUSDC = "USDC"`, and `DefaultQuoteExpiry = 5 * time.Minute` as the single source of truth.

#### Scenario: Constants are accessible
- **WHEN** code imports `internal/finance`
- **THEN** `finance.USDCDecimals`, `finance.CurrencyUSDC`, and `finance.DefaultQuoteExpiry` are available

### Requirement: wallet package re-exports finance functions
The `wallet` package SHALL re-export `ParseUSDC`, `FormatUSDC`, `CurrencyUSDC`, and `USDCDecimals` by delegating to the `finance` package, maintaining backward compatibility for existing callers.

#### Scenario: wallet.ParseUSDC delegates to finance
- **WHEN** `wallet.ParseUSDC("1.00")` is called
- **THEN** the result is identical to `finance.ParseUSDC("1.00")`

### Requirement: finance package has zero internal dependencies
The `internal/finance` package MUST NOT import any other `internal/` package. It is a leaf package.

#### Scenario: Import graph is clean
- **WHEN** the import graph of `internal/finance` is inspected
- **THEN** no imports from `github.com/langoai/lango/internal/` are present
