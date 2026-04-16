# Tasks: fix-extension-installer-security

- [x] Fix 1: Add rootDir param to copyTree, per-file ResolvePath in Walk callback
- [x] Fix 1: Add os.Lstat symlink rejection in copyFile
- [x] Fix 1: Update callers (copyPackFiles, copySkillsToStore) to pass rootDir
- [x] Fix 1: Add TestCopyTreeRejectsSymlinkEscape + TestCopyFileRejectsSymlink
- [x] Fix 2: Add looksLikeSHA helper and cloneAndCheckout helper
- [x] Fix 2: Update GitSource.Fetch to use SHA-aware clone+checkout strategy
- [x] Fix 2: Add TestLooksLikeSHA table test + TestGitSourceFetchSHA with local repo fixture
- [x] Fix 3: Update plannedWrites to accept rootDir and walk skill directories
- [x] Fix 3: Update Inspect call site to pass wc.RootDir
- [x] Fix 3: Add TestPlannedWritesIncludesDirectoryContents
- [x] Verify: go test ./internal/extension/... — ALL PASS
