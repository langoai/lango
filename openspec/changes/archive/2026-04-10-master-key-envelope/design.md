## Context

Lango의 현재 로컬 저장 암호화 구조는 `internal/security/local_provider.go`에서 `passphrase → PBKDF2(100k, SHA256) → AES-256-GCM key`로 key를 직접 유도한다. 이 key가 `secrets` 테이블과 `config_profiles` 테이블 암호화, 그리고 `session/ent_store.go`를 통해 `~/.lango/lango.db`의 SQLCipher key 역할까지 수행한다.

이 구조는 단순하지만 3가지 구조적 문제가 있다:
1. **Passphrase 분실 = 전체 상실**: recovery 경로 없음
2. **Passphrase 변경 = 전체 재암호화**: 모든 secrets, config_profiles를 새 key로 재암호화
3. **Storage root와 identity root 혼합**: Phase 1 이후 Phase 2-4 (algorithm agility, DID v2, hybrid handshake)에서 identity 분리 시 storage 쪽이 걸림돌

또한 현재 `bootstrap.go:116`과 `session/ent_store.go:99-102` 사이에 SQLCipher key 전달 방식이 불일치한다 (`PRAGMA key = '<passphrase>'` vs `PRAGMA key = "x'<hex>'"`). 이번 change에서 raw key mode로 통일한다.

**Stakeholders**: CLI 사용자 (passphrase 변경/복구 UX), CryptoProvider consumer (인터페이스 불변), 향후 Phase 2-7 PQC 작업.

## Goals / Non-Goals

**Goals:**
- Random Master Key (MK) 생성, 여러 KEK slot으로 wrap/unwrap 지원
- Passphrase 변경 시 data 재암호화 불필요 (envelope re-wrap only)
- BIP39 24-word recovery mnemonic slot 지원
- DB encryption key = HKDF(MK) 로 passphrase와 완전 분리
- Migration 중 crash 발생 시 안전한 복구 (PendingMigration + PendingRekey flags)
- CryptoProvider 인터페이스 불변 — 모든 consumer 무영향
- `lango security status` passphrase-free 기본 동작 (spec 준수)
- 9개 OpenSpec spec에 delta 반영

**Non-Goals:**
- PQC 알고리즘 도입 (Phase 2-3 대상)
- Identity root 분리 (Phase 0, 3 대상)
- Hardware token slot 구현 (`KEKSlotHardware` 타입만 정의, 실제 구현은 follow-up)
- `recovery_file` slot 구현 (`KEKSlotRecoveryFile` 타입만 정의, 실제 구현은 follow-up)
- Argon2id KDF 지원 (KDFAlg metadata 필드만 준비, 실제 구현은 follow-up)
- `LangoDir`을 `DBPath` 부모 디렉토리로 연동 (Phase 1 범위 밖)
- Mnemonic recovery 자동화 / headless 환경 지원 (interactive만)

## Decisions

### D1: Envelope 저장 위치 — 파일시스템 (`~/.lango/envelope.json`)

**선택**: `~/.lango/envelope.json` (JSON, 0600 permissions)

**대안**:
- (A) DB 안 `security_config` 테이블 row — SQLCipher 활성 시 passphrase 분실하면 DB 자체를 열 수 없어 envelope 접근 불가. mnemonic recovery 불가능.
- (B) OS keyring — 플랫폼 의존성 증가, 테스트 어려움.

**선택 이유**: (A)의 SQLCipher + recovery 경로 모순을 해결하는 유일한 방법은 envelope을 DB 외부에 두는 것. JSON 파일은 WrappedMK가 KEK로 암호화돼 있어 plaintext로 두어도 안전 (LUKS, KeePass와 동일 설계).

### D2: DB key 유도 — `HKDF(MK, "lango-db-encryption")` raw 32-byte

**선택**: `DeriveDBKey(mk) = HKDF-Expand(SHA256, mk, "lango-db-encryption", 32)`, hex 인코딩 후 `PRAGMA key = "x'<hex>'"`

**대안**:
- (A) Passphrase 직접 사용 (현재) — passphrase 변경 시 `PRAGMA rekey` 필요.
- (B) 별도 random DB key 생성 후 MK로 wrap — envelope에 추가 필드 필요, 구현 복잡.
- (C) `HKDF(MK, ...)` — MK 불변이면 DB key 불변, passphrase 변경과 완전 독립.

**선택 이유**: (C)가 가장 단순. MK가 유일한 루트이므로 MK에서 파생되는 모든 key는 자동으로 안정. Passphrase 변경 = envelope re-wrap only, DB 영향 없음.

### D3: Migration crash safety — `PendingMigration` + `PendingRekey` dual-flag

**선택**: Envelope에 `PendingMigration`과 `PendingRekey` 2개 boolean flag. Envelope을 data re-encryption 전에 먼저 저장 (flag=true). 각 단계 완료 시 flag 해제.

**대안**:
- (A) 단일 flag — rekey 실패 시 상태 구분 불가.
- (B) SQL transaction만 사용 — `PRAGMA rekey`는 transactional이 아님.
- (C) 별도 "migration_state" 파일 — 중복 상태 관리.

**선택 이유**: 2개 flag로 "data migration done, rekey pending" 중간 상태를 정확히 표현. Bootstrap Phase 7이 flag를 직접 읽고 재시도. Plaintext DB도 PendingMigration을 사용해 동일 경로로 처리.

### D4: DB 백업 — `PRAGMA wal_checkpoint(TRUNCATE)` + `VACUUM INTO`

**선택**: `PRAGMA wal_checkpoint(TRUNCATE)` 후 `VACUUM INTO 'lango.db.pre-migration'`

**대안**:
- (A) `cp lango.db lango.db.pre-rekey` — WAL 모드에서 `-wal` 파일 내용 누락 위험.
- (B) SQLite backup API — CGO 호출, 복잡.
- (C) `VACUUM INTO` — WAL 상태 무관하게 일관된 스냅샷 생성.

**선택 이유**: (C)가 WAL-safe하면서 SQL 레벨에서 동작. `wal_checkpoint` 선행으로 추가 안전성 확보.

### D5: Bootstrap pipeline — 7 phases → 10 phases

**선택**:
```
1. EnsureDataDir (수정)  2. DetectEncryption  3. LoadEnvelopeFile (신규)
4. AcquireCredential (수정)  5. UnwrapOrCreateMK (신규)  6. OpenDatabase (수정)
7. MigrateEnvelope (신규)  8. LoadSecurityState (수정)  9. InitCrypto (수정)
10. LoadProfile
```

**대안**:
- (A) 기존 7 phases 유지하고 phase 내부에서 분기 — phase 책임이 흐려짐, 테스트 어려움.
- (B) 단일 거대 phase — 단순하지만 cleanup/재시작 관리 어려움.

**선택 이유**: 각 phase가 단일 책임. LoadEnvelopeFile이 DB 접근 전에 실행되어 recovery 옵션 판단 가능. MigrateEnvelope가 별도 phase로 분리되어 crash 재시도 로직이 명확.

### D6: CryptoProvider 인터페이스 불변

**선택**: `Sign/Encrypt/Decrypt` 시그니처 변경 없음. `LocalCryptoProvider`에 신규 메서드 `InitializeWithEnvelope`, `InitializeNewEnvelope`, `Envelope()`, `IsLegacy()`, `Close()`만 추가.

**대안**:
- (A) 인터페이스에 KEK 관리 메서드 추가 — 모든 구현체 (KMS, Composite, RPC) 영향.
- (B) 새 인터페이스 `EnvelopeCryptoProvider` 정의 — wiring 복잡, 기존 consumer 변경.

**선택 이유**: Consumer 안정성이 최우선. MK는 `keys["local"]`에 저장되어 기존 Encrypt/Decrypt/Sign 경로 그대로 사용. KMS provider들은 이 change와 무관.

### D7: BIP39 라이브러리 — `tyler-smith/go-bip39`

**선택**: `github.com/tyler-smith/go-bip39` (사용자 확정)

**대안**:
- (A) `cosmos/go-bip39` — Cosmos 생태계 지향, 과도.
- (B) 직접 구현 — wordlist 관리 부담, 버그 위험.

**선택 이유**: tyler-smith/go-bip39는 Go 생태계 표준. 가볍고 단일 목적. go-ethereum과 독립적.

### D8: `security_config` 접근 통합 — `SecurityConfigStore`

**선택**: `internal/security/config_store.go`에 `SecurityConfigStore` struct. bootstrap.go와 session/ent_store.go의 중복 `ensureSecurityTable`/`loadSalt`/`storeSalt`/`loadChecksum`/`storeChecksum`을 이 struct로 위임.

**대안**:
- (A) 중복 유지 — drift 위험 증가.
- (B) Ent entity로 모델링 — 기존 raw SQL 패턴과 충돌.

**선택 이유**: 현재 2곳에 중복된 코드가 이번 change에서 3곳 이상으로 늘어날 위험. 단일 store로 수렴하는 것이 유지보수 측면에서 정확.

### D9: `lango security status` 기본 동작 — passphrase-free non-interactive mini-bootstrap

**선택**: Envelope 파일 직접 읽기 + `passphrase.AcquireNonInteractive()` (keyring/keyfile only, prompt 없음) + `openDatabaseReadOnly()` (read-only, no schema migration).

**대안**:
- (A) 기존 `bootLoader()` 호출 — interactive 환경에서 prompt 띄움, spec 위반.
- (B) `--full` 플래그로만 DB 접근 — 기존 스펙의 "default passphrase-free" 요구사항 위반.

**선택 이유**: 현재 `cli-security-status` spec은 기본 동작이 passphrase 없이 기존 필드를 모두 보여야 하고, DB 접근 불가 시 zero counts로 degrade해야 함. 이를 준수하려면 non-interactive mini-bootstrap이 필수. 최근 커밋 `8112bcc6`의 sandbox graceful degradation 패턴 참조.

### D10: `openDatabase()` 시그니처에 `rawKey bool` 추가

**선택**: `openDatabase(dbPath, encryptionKey string, rawKey bool, cipherPageSize int)`

**대안**:
- (A) Raw key 전용 openDatabase 분리 — 두 함수 유지 부담.
- (B) `encryptionKey`에 prefix (`"raw:..."`) — fragile, 에러 취약.

**선택 이유**: `rawKey bool` 파라미터가 명확. 같은 코드 경로에서 `PRAGMA key = '...'` vs `PRAGMA key = "x'...'"` 분기만 처리.

## Risks / Trade-offs

- **[Critical] `PRAGMA rekey` 실패 → split state** → `PendingRekey` flag + dual-open fallback (Phase 6: MK-derived key 실패 시 legacy passphrase fallback) + `VACUUM INTO` 백업 보존
- **[Critical] MK가 log/error 메시지에 노출** → MK는 error string에 절대 포함 금지. `fmt.Errorf("... %w", err)` wrapping만 사용. 코드 리뷰에서 grep으로 검증
- **[High] Migration data re-encrypt 중 crash (plaintext)** → `PendingMigration` flag로 다음 부팅 시 재시도
- **[High] Migration TX + rekey 사이 crash (SQLCipher)** → PendingMigration=false + PendingRekey=true 상태로 저장. 다음 부팅 시 passphrase로 DB open (legacy key) → rekey 재시도
- **[High] `config_profiles` re-encryption 누락** → Migration TX 안에서 COUNT 검증 (전후 행 수 일치 필수)
- **[Medium] Envelope file 퍼미션 문제** → 0600으로 생성. 로딩 시 퍼미션 검증 경고 (warning만, 강제 거부 안 함)
- **[Medium] **Residual**: PendingMigration/PendingRekey=true 상태 + passphrase 분실 + mnemonic만 존재** → mnemonic으로 MK unwrap은 가능하지만 DB가 legacy passphrase key이므로 열 수 없음. `VACUUM INTO` 백업(`lango.db.pre-migration`)에서 복원 필요. 이 시나리오는 (migration crash) ∩ (passphrase 분실)이므로 극히 드묾. 문서 + CLI 경고에 명시
- **[Low] BIP39 dependency 취약점** → `tyler-smith/go-bip39` 버전 고정, 간접 의존성 감사
- **[Low] 기존 `migrate-passphrase` 사용자 혼란** → deprecation 메시지 + `change-passphrase` 안내
- **[Low] PBKDF2 파라미터 업그레이드 (향후)** → `KEKSlot.KDFParams`에 iteration 저장. 향후 새 slot 추가 시 더 높은 값 사용 가능, 기존 slot은 그대로 유지

## Migration Plan

### Phase 1a: 신규 설치 (first-run)
- `EnsureDataDir` → `LoadEnvelopeFile` (nil) → `AcquireCredential` (passphrase) → `UnwrapOrCreateMK` (신규 envelope 생성) → `OpenDatabase` (MK-derived key 또는 plaintext) → `InitCrypto`
- 사용자 관점: 기존과 동일한 passphrase 프롬프트

### Phase 1b: 기존 설치 업그레이드 (legacy → envelope)
- `EnsureDataDir` → `LoadEnvelopeFile` (nil) → `AcquireCredential` (passphrase) → `UnwrapOrCreateMK` (LegacyMode=true, skip) → `OpenDatabase` (passphrase as legacy key) → `MigrateEnvelope`:
  1. Derive old key + checksum 검증
  2. MK 생성, envelope 생성 (PendingMigration=true, SQLCipher면 PendingRekey=true)
  3. `StoreEnvelopeFile()` — envelope 먼저 저장
  4. `wal_checkpoint(TRUNCATE)` + `VACUUM INTO 'lango.db.pre-migration'` — WAL-safe 백업
  5. SQL TX: secrets + config_profiles re-encrypt (COUNT 검증 포함)
  6. PendingMigration=false, envelope 갱신
  7. (SQLCipher) `PRAGMA rekey = "x'<HKDF(MK)>'"` + close/reopen 검증
  8. PendingRekey=false, envelope 갱신
  9. (선택) 백업 파일 제거
- 사용자 관점: 첫 부팅 시 "Upgrading encryption format..." 메시지. 이후 정상 동작.

### Phase 1c: Crash 복구
- Crash 지점별 복구는 Bootstrap Phase 7 (MigrateEnvelope)에서 `s.Envelope.PendingMigration`/`PendingRekey`를 직접 읽어 재시도.
- 다음 부팅 시 자동으로 완료됨.

### Rollback
- Phase 1의 구현 자체는 rollback하지 않음 (downgrade 시 기존 legacy 경로가 envelope을 무시할 수 없음).
- 긴급 상황 시 `lango.db.pre-migration` 백업 파일을 `lango.db`로 복원 + `envelope.json` 삭제 → 기존 binary로 동작.
- 문서화 필요: `docs/security/envelope-migration.md`에 백업/복원 절차 명시.

## Open Questions

- **Q1**: `SecurityConfigStore` 통합 범위 — bootstrap.go와 session/ent_store.go 중복만 정리할지, 아니면 `MigrateSecrets`까지 포함할지? **답**: 이번 change는 중복 제거까지만. `MigrateSecrets`는 내부 로직이 있어 별도 follow-up.
- **Q2**: Envelope file 퍼미션 검증 — 0600이 아닌 경우 reject할지 warning만 낼지? **답**: Phase 1에서는 warning만. reject은 tighter security posture가 필요할 때 follow-up.
- **Q3**: Recovery mnemonic의 passphrase 보호 (BIP39 optional passphrase) — Phase 1에서 지원할지? **답**: 지원 안 함. BIP39 optional passphrase는 UX 복잡도 증가. follow-up.
