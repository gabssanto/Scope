# Scope CLI Roadmap

A phased implementation plan from easiest to most complex features.

---

## Phase 1: Quick Wins (Low Effort, High Value) ✅ COMPLETE

Simple commands that build on existing infrastructure.

### 1.1 `scope rename <old> <new>` ✅
Rename a tag across all folders.

**Implementation:**
- Single SQL UPDATE on `tags` table
- Validate new name doesn't exist
- ~30 lines of code

**Files:** `cmd/scope/main.go`, `internal/tag/manager.go`

---

### 1.2 `scope tags <path>` ✅
Show all tags for a specific folder (inverse of `scope list <tag>`).

**Implementation:**
- Already have `GetTagsForFolder()` in manager
- Just wire up CLI command
- ~15 lines of code

**Files:** `cmd/scope/main.go`

---

### 1.3 `scope prune` ✅
Remove tags pointing to deleted/non-existent folders.

**Implementation:**
- Query all folders, check `os.Stat()` for each
- Delete rows where path doesn't exist
- Add `--dry-run` flag to preview
- ~40 lines of code

**Files:** `cmd/scope/main.go`, `internal/tag/manager.go`

---

### 1.4 `scope debug` ✅
Print debug info for troubleshooting.

**Implementation:**
- Print: version, db path, config path, OS/arch, shell
- Useful for bug reports
- ~25 lines of code

**Files:** `cmd/scope/main.go`

---

## Phase 2: Navigation Commands ✅ COMPLETE (except pick)

Commands that improve daily workflow.

### 2.1 `scope go <tag>` ✅
Quick jump to a tagged folder.

**Implementation:**
- If single folder: print path (shell wrapper does `cd`)
- If multiple: show numbered list, prompt for selection
- Output path to stdout for shell integration
- ~50 lines of code

**Shell integration (user adds to .bashrc/.zshrc):**
```bash
sg() { cd "$(scope go "$@")" }
```

**Files:** `cmd/scope/main.go`

---

### 2.2 `scope pick [tag]`
Interactive fuzzy finder.

**Implementation:**
- List all folders (or filtered by tag)
- Simple built-in picker with arrow keys
- Optional: detect if `fzf` installed and use it
- ~80 lines of code

**Dependencies:** None (can use simple stdin picker)

**Files:** `cmd/scope/main.go`, `internal/picker/picker.go` (new)

---

### 2.3 `scope open <tag>` ✅
Open folder(s) in system file manager.

**Implementation:**
- macOS: `open <path>`
- Linux: `xdg-open <path>`
- Windows: `explorer <path>`
- ~30 lines of code

**Files:** `cmd/scope/main.go`

---

### 2.4 `scope edit <tag>` ✅
Open folder(s) in editor.

**Implementation:**
- Use `$EDITOR` or `$VISUAL` or fallback to `code`/`vim`
- If multiple folders, open each
- ~30 lines of code

**Files:** `cmd/scope/main.go`

---

## Phase 3: Tag Management

Better tag organization capabilities.

### 3.1 `scope merge <tag1> <tag2>`
Merge tag1 into tag2.

**Implementation:**
- Move all folder associations from tag1 to tag2
- Delete tag1
- Handle duplicates (folder already has tag2)
- ~40 lines of code

**Files:** `cmd/scope/main.go`, `internal/tag/manager.go`

---

### 3.2 `scope clone <tag> <new-tag>`
Copy tag associations to new tag.

**Implementation:**
- Get all folders for tag
- Add new tag to each folder
- ~30 lines of code

**Files:** `cmd/scope/main.go`, `internal/tag/manager.go`

---

### 3.3 `scope alias <name> <tags...>`
Create tag groups.

**Implementation:**
- New table `tag_aliases` (name, tags JSON array)
- Modify `start` to expand aliases
- ~60 lines of code

**Files:** `internal/db/sqlite.go`, `internal/tag/manager.go`, `cmd/scope/main.go`

---

## Phase 4: Bulk Operations ✅ MOSTLY COMPLETE

Commands that work across multiple folders.

### 4.1 `scope each <tag> <cmd>` ✅
Run command in each tagged folder.

**Implementation:**
- Get folders for tag
- For each: `cd` to folder, run command, capture output
- Flags: `--parallel` / `-p` for concurrent execution
- Show folder name before each output
- ~100 lines of code

**Example:**
```bash
scope each work "git status -s"
scope each backend -p "go test ./..."
```

**Files:** `cmd/scope/main.go`, `internal/exec/runner.go` (new)

---

### 4.2 `scope status <tag>` ✅
Git status across tagged folders (shortcut).

**Implementation:**
- Wrapper around `scope each <tag> "git status -s"`
- Only show folders with changes
- ~30 lines of code

**Files:** `cmd/scope/main.go`

---

### 4.3 `scope pull <tag>` ✅
Git pull across tagged folders.

**Implementation:**
- Wrapper around `scope each <tag> -p "git pull"`
- Show success/failure summary
- ~30 lines of code

**Files:** `cmd/scope/main.go`

---

### 4.4 `scope scan <dir>` ✅ (Already implemented)
Auto-detect and tag projects via `.scope` files.

**Implementation:**
- Reads `.scope` YAML files with `tags:` array
- Interactive UI with charmbracelet/huh
- Already in `internal/scan/` package

---

## Phase 5: Global Configuration

Support for global config file (`.scope` project files already implemented).

### 5.1 Global Config File
`~/.config/scope/config.yml`

**Implementation:**
```yaml
# ~/.config/scope/config.yml
editor: code
update:
  check: true
  interval: 24h
scan:
  depth: 3
  auto_tags:
    - type: nodejs
      tags: [js, frontend]
```

- Load config on startup
- `scope config` to view
- `scope config set <key> <value>` to modify
- ~100 lines of code

**Files:** `internal/config/config.go` (new), `cmd/scope/main.go`

---

> **Note:** Project-level `.scope` files are already implemented via `scope scan`.

---

## Phase 6: Self-Update System ✅ COMPLETE

### 6.1 Update Check (Background) ✅
Check for new versions without blocking.

**Implementation:**
- On command run, async check GitHub releases API
- Cache result in `~/.config/scope/.update-check`
- Only check once per 24h
- Only show notice on interactive TTY
- ~80 lines of code

**Files:** `internal/update/checker.go` (new), `cmd/scope/main.go`

---

### 6.2 `scope update` ✅
Self-update command.

**Implementation:**
- Fetch latest release from GitHub API
- Compare with current version
- Download appropriate binary (OS/arch detection)
- Verify checksum
- Replace current binary (handle permissions)
- ~150 lines of code

**Flags:**
- `--check` - only check, don't update
- `--force` - update even if current
- `--version <ver>` - install specific version

**Files:** `internal/update/updater.go` (new), `cmd/scope/main.go`

---

### 6.3 `scope changelog`
Show what's new.

**Implementation:**
- Fetch CHANGELOG.md from repo or embed at build time
- Show entries newer than current version
- ~40 lines of code

**Files:** `cmd/scope/main.go`, `internal/update/changelog.go` (new)

---

## Phase 7: Polish & DX

### 7.1 `scope completions <shell>`
Shell completion scripts.

**Implementation:**
- Generate completion scripts for bash, zsh, fish
- Complete tag names dynamically
- ~200 lines of code (templates for each shell)

**Usage:**
```bash
# Add to .bashrc
eval "$(scope completions bash)"
```

**Files:** `internal/completions/completions.go` (new)

---

### 7.2 `scope doctor`
Health check and diagnostics.

**Implementation:**
- Check db exists and is readable
- Check for broken paths (folders that don't exist)
- Check permissions
- Check for orphaned tags
- Suggest fixes
- ~80 lines of code

**Files:** `cmd/scope/main.go`, `internal/doctor/doctor.go` (new)

---

### 7.3 `scope export` / `scope import`
Backup and restore.

**Implementation:**
```yaml
# scope export > backup.yml
version: 1
tags:
  work:
    - /path/to/project1
    - /path/to/project2
  personal:
    - /path/to/blog
```

- Export: dump all tags and folders to YAML
- Import: read YAML and create tags (with merge strategy)
- Flags: `--format json|yaml`, `--merge`
- ~100 lines of code

**Files:** `cmd/scope/main.go`, `internal/export/export.go` (new)

---

## Phase 8: Advanced Features

### 8.1 `scope clone <git-url> [tag]`
Git clone with auto-tagging.

**Implementation:**
- Clone repo to current dir (or configured projects dir)
- Auto-detect project type
- Apply tags (provided + auto-detected)
- ~60 lines of code

**Files:** `cmd/scope/main.go`

---

### 8.2 `scope watch <dir>`
Watch directory for new projects.

**Implementation:**
- Use fsnotify for filesystem events
- When new directory created, run auto-detection
- Optionally auto-tag
- Run as daemon or foreground
- ~120 lines of code

**Files:** `internal/watch/watcher.go` (new)

**Dependencies:** `github.com/fsnotify/fsnotify`

---

### 8.3 Multi-tag Sessions
`scope start <tag1> <tag2> ...`

**Implementation:**
- Modify start to accept multiple tags
- Union of all folders
- Handle duplicates
- ~40 lines of code

**Files:** `internal/session/scope.go`, `cmd/scope/main.go`

---

### 8.4 Session Hooks
Run commands on session start/exit.

**Implementation:**
```yaml
# config.yml
hooks:
  on_start:
    - tmux rename-window "scope:${SCOPE_SESSION}"
  on_exit:
    - echo "Exited scope session"
```

- Execute hooks in order
- Pass environment variables
- ~60 lines of code

**Files:** `internal/session/scope.go`, `internal/config/config.go`

---

## Implementation Order Summary

| Phase | Commands | Status |
|-------|----------|--------|
| 1 | rename, tags, prune, debug | ✅ Complete |
| 2 | go, pick, open, edit | ✅ Complete (pick pending) |
| 3 | merge, clone, alias | Pending |
| 4 | each, status, pull, scan | ✅ Complete |
| 5 | config, .scope.yml | Pending |
| 6 | update check, update, changelog | ✅ Complete (changelog pending) |
| 7 | completions, doctor, export/import | Pending |
| 8 | git clone, watch, multi-tag, hooks | Pending |

### Implemented Commands (Current Session)
- `scope rename <old> <new>` - Rename tags
- `scope tags <path>` - Show tags for a folder
- `scope prune [--dry-run]` - Remove stale folders
- `scope debug` - Debug information
- `scope go <tag>` - Quick jump to folder
- `scope open <tag>` - Open in file manager
- `scope edit <tag>` - Open in editor
- `scope each <tag> <cmd>` - Run command in each folder
- `scope status <tag>` - Git status across folders
- `scope pull <tag>` - Git pull across folders
- `scope update [--check]` - Self-update with version check

---

## Next Steps

1. Start with Phase 1 (quick wins)
2. Get user feedback
3. Iterate through phases
4. Release incrementally (v0.2, v0.3, etc.)

---

## Version Planning

- **v0.2.0** - Phase 1 + 2 (navigation essentials)
- **v0.3.0** - Phase 3 + 4 (bulk operations, scan)
- **v0.4.0** - Phase 5 + 6 (config, self-update)
- **v0.5.0** - Phase 7 (polish)
- **v1.0.0** - Phase 8 + stability
