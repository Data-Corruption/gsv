# Changelog

## [v0.1.3] - 2025-08-22

### Fixed
- main: Changed return logic. I read os.Exit ignores defers, unsure if true and too lazy to test so i'm playing it safe.

## [v0.1.2] - 2025-08-13

### Fixed
- main: Removed mistaken additional part in log path.

### Changed
- readme: simplified and added build status badge.
- workflow: small rename

## [v0.1.1] - 2025-08-13

### Fixed
- ascon: Correctly flush previous vector on new `Count`; prevents dropping vectors during parse.
- main: Remove panics during startup; exit non‑zero with clear stderr messages instead.

### Changed
- ascon: Thread `context.Context` through parser/generator; promote key logs to info; add scan diagnostics.
- main: Use `Name` for log dir (`~/.[name]/logs`) and `DefaultLogLevel` for logger init.
- gitignore: Add `test/` to `.gitignore`.
 - main: Add Ctrl‑C handling; respect missing HOME; write errors to stderr; non‑zero exit on failure.
 - update: Respect global `--yes` to auto‑approve sudo prompts; better error messages; wire timeouts to parent context.
 - parser: Increase scanner buffer to support large KAT lines.

## [v0.1.0] - 2025-08-11

### Added
- gsv CLI with global flags: `--log` (log level) and `-y/--yes` (assume yes).
- Generator: `ascon` — converts Ascon AEAD KAT (.rsp) files into a SystemVerilog ROM.
  - Default output: `ascon_rom.sv`
  - Byte order option: `--swap-bytes`/`--sb` (data bytes only; padding unchanged)
  - Safe writes via temp file and overwrite prompt (skipped with `-yes`)
- Command: `update` — self‑updates to the latest GitHub release, prompting for sudo when needed.
- Install scripts for Linux and Windows (WSL).
- Structured logging to `~/.gsvr/logs` with default level `warn`.