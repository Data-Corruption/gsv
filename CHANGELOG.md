# Changelog

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