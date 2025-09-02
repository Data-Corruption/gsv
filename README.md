# gsvc [![Build](https://github.com/Data-Corruption/gsvc/actions/workflows/build.yml/badge.svg)](https://github.com/Data-Corruption/gsvc/actions/workflows/build.yml)

Collection of miscellaneous SystemVerilog code generators.

## Install

> _Needs: x64, Linux or WSL_

### Linux

```sh
curl -sSfL https://raw.githubusercontent.com/Data-Corruption/gsvc/main/scripts/install.sh | bash
```

### Windows (with WSL installed)

PowerShell:
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; iex "& { $(irm https://raw.githubusercontent.com/Data-Corruption/gsvc/main/scripts/install.ps1) }"
```

## Usage

`gsvc subcommand <required arg> [optional arg]`  
`gsvc -h subcommand` Help for any subcommand

### ascon

```sh
gsvc ascon <inPath> [outPath] [--sb]
```

Reads a 128-bit Known-Answer Test (KAT) vector file ([example](https://github.com/ascon/ascon-c/blob/main/crypto_aead/asconaead128/LWC_AEAD_KAT_128_128.txt)) and produces a SV ROM.
`--sb` swaps byte order of data portion (padding unchanged).

### update

```sh
gsvc update
```
> updates gsvc to the latest release.
