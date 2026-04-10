## Why

현재 Lango의 로컬 저장 암호화는 `passphrase → PBKDF2 → AES-256-GCM key` 구조로, passphrase가 data key와 SQLCipher DB key 양쪽 역할을 직접 수행한다. 이로 인해 (1) passphrase 분실 시 전체 데이터 상실, (2) passphrase 변경 시 전체 재암호화 필요, (3) 복구 경로 부재라는 3가지 구조적 문제가 있다. 이는 Phase 1 Security & Crypto Renewal의 첫 단계로, storage root를 identity root와 분리하고 향후 PQC 전환을 위한 key hierarchy 기반을 마련한다.

## What Changes

- **신규**: Master Key Envelope (MK/KEK) 아키텍처 도입
  - 랜덤 32바이트 Master Key (MK) 생성, AES-256-GCM wrapped 상태로 `~/.lango/envelope.json`에 저장
  - Passphrase → KEK (PBKDF2) → MK unwrap 3-layer hierarchy
  - KEK slot 모델로 passphrase + recovery mnemonic slot 동시 지원
  - KEK slot에 KDF metadata (`KDFAlg`, `KDFParams`, `WrapAlg`, `Domain`) 포함하여 향후 알고리즘 전환 가능
- **신규**: BIP39 기반 recovery mnemonic (24-word)
  - `lango security recovery setup` — mnemonic 생성 + slot 추가
  - `lango security recovery restore` — mnemonic으로 MK unwrap + 새 passphrase 설정
- **신규**: `internal/security/config_store.go` — `security_config` 테이블 접근 통합 (bootstrap, session store 중복 제거)
- **수정 (BREAKING internal)**: Bootstrap pipeline 7 → 10 phases
  - 신규 phase: LoadEnvelopeFile, UnwrapOrCreateMK, MigrateEnvelope
  - Phase 순서 역전: envelope 로드 → passphrase → MK unwrap → DB open
  - `phaseEnsureDataDir`이 `Options.LangoDir` 우선 적용
- **수정**: DB encryption key = HKDF(MK, "lango-db-encryption") (raw 32-byte key, `PRAGMA key = "x'<hex>'"`)
  - Passphrase 변경 시 DB rekey 불필요 (MK 불변)
  - Legacy 설치 시 1회 migration + `PRAGMA rekey` (WAL-safe `VACUUM INTO` 백업)
- **수정**: `lango security change-passphrase` 신규 (envelope re-wrap only, 데이터 재암호화 없음)
  - `migrate-passphrase` deprecated
- **수정**: `lango security status` 기본 동작을 passphrase-free로 전환 (keyring/keyfile non-interactive, graceful degradation)
- **수정**: `openDatabase()` 시그니처에 `rawKey bool` 추가 (`PRAGMA key = "x'<hex>'"` vs `'<passphrase>'`)
- **추가**: Crash recovery — `PendingMigration` + `PendingRekey` flags로 migration 중 crash 흡수
- **CryptoProvider 인터페이스는 불변** — 모든 consumer (SecretsStore, ConfigStore, KMS providers, tools) 영향 없음

## Capabilities

### New Capabilities

- `master-key-envelope`: MK/KEK 3-layer key hierarchy. Envelope 파일 저장, KEK slot 모델, passphrase/mnemonic slot, KDF metadata, crash recovery flags, MK wrap/unwrap 계약
- `recovery-mnemonic`: BIP39 24-word mnemonic 기반 recovery. 생성, 검증, slot 추가, restore flow

### Modified Capabilities

- `passphrase-management`: passphrase가 data key에서 KEK로 역할 전환. `migrate-passphrase` deprecated, `change-passphrase` 신규 (envelope re-wrap)
- `passphrase-acquisition`: envelope 존재 시 recovery mnemonic 선택 옵션 추가. `AcquireNonInteractive()` 신규 (keyring/keyfile만, prompt 없음)
- `bootstrap-lifecycle`: phase 수 7→10, 순서 역전 (envelope 먼저, DB 나중), 신규 phase 3개, `Options.LangoDir` 추가
- `db-encryption`: DB key derivation = HKDF(MK), raw key mode (`"x'<hex>'"`), 1회 `PRAGMA rekey` migration, WAL-safe backup
- `cli-security-status`: passphrase-free 기본 동작, envelope 섹션 추가, non-interactive mini-bootstrap, graceful degradation

> Note: `encrypted-config-profiles`, `cli-secrets-management`, `keyfile-shred`, `keyring-security-tiering`는 CryptoProvider 인터페이스가 불변이므로 spec-level 변경 없음 (내부 구현만 변경).

## Impact

### Code

- **신규 파일 (10)**: `internal/security/envelope.go`, `envelope_file.go`, `mnemonic.go`, `config_store.go`, `migrate_envelope.go` (+ tests), `internal/cli/security/change_passphrase.go`, `recovery.go`
- **수정 파일 (8)**: `internal/security/local_provider.go`, `errors.go`, `internal/bootstrap/bootstrap.go`, `phases.go`, `pipeline.go`, `internal/cli/security/status.go`, `migrate.go`, `internal/session/ent_store.go`
- **변경 없음**: `crypto.go` (인터페이스), `secrets_store.go`, `key_registry.go`, `configstore/store.go`, `composite_provider.go`, 모든 KMS provider, wallet, P2P

### Dependencies

- **신규**: `github.com/tyler-smith/go-bip39` (BIP39 mnemonic 라이브러리)
- **기존 활용**: `golang.org/x/crypto/pbkdf2`, `golang.org/x/crypto/hkdf`

### Data / Storage

- **신규 파일**: `~/.lango/envelope.json` (0600 permissions, JSON encoded)
- **DB 영향**: 1회 migration 시 `secrets`, `config_profiles` 전체 re-encryption + (SQLCipher 한정) `PRAGMA rekey`
- **스키마 변경 없음**: 기존 ent 엔티티 (`Key`, `Secret`, `ConfigProfile`) 그대로 유지

### User-facing

- **새 명령**: `lango security change-passphrase`, `lango security recovery setup`, `lango security recovery restore`
- **Deprecated**: `lango security migrate-passphrase`
- **동작 변경**: `lango security status` — passphrase prompt 없이 기본 동작, envelope 정보 표시
- **Migration UX**: 기존 설치 첫 부팅 시 "Upgrading encryption format (one-time migration)..." 메시지 표시

### Residual Risk

- `PendingMigration=true` 또는 `PendingRekey=true` 상태에서 passphrase를 잃으면, mnemonic으로 MK unwrap은 가능하지만 DB가 legacy passphrase key이므로 열 수 없음. 복구는 `VACUUM INTO` 백업(`lango.db.pre-migration`)에서 복원 필요. 이는 (migration crash) ∩ (passphrase 분실) 교집합이므로 극히 드문 시나리오이나 문서 및 CLI 경고에 명시.
