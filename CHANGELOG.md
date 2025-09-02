# Changelog

## [v0.1.1] - 2025-09-02

### Changed
- renamed from gsv to gsvc due to gsv being reserved in Windows.

## [v0.1.0] - 2025-09-02

### Added
- gsvc CLI with global flags: `--log` (log level) and `-y/--yes` (assume yes).
- Generator: `ascon` — converts Ascon AEAD KAT (.rsp) files into a SystemVerilog ROM.
  - Default output: `ascon_rom.sv`
  - Byte order option: `--swap-bytes`/`--sb` (data bytes only; padding unchanged)
  - Safe writes via temp file and overwrite prompt (skipped with `-yes`)
- Command: `update` — self‑updates to the latest GitHub release, prompting for sudo when needed.
- Install scripts for Linux and Windows (WSL).
- Structured logging to `~/.gsvc/logs` with default level `warn`.