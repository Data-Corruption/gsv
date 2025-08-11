# gsv

Collection of miscellaneous SystemVerilog code generators.

## Installation

Prerequisites:
- x64 machine running linux or wsl
- that's it :p

### Linux

```sh
curl -sSfL https://raw.githubusercontent.com/Data-Corruption/gsv/main/scripts/install.sh | bash
```

### Windows With WSL

powershell (as administrator):
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; iex "& { $(irm https://raw.githubusercontent.com/Data-Corruption/gsv/main/scripts/install.ps1) }"
```

## Generators

Usage format: `gsv subcommand <required-arg> [optional-arg]`.  
For more info on any given subcommand: `gsv -h subcommand_name`.

### ascon

`gsv ascon <in-path> [out-path] [--sb]`

Ingests a 128 bit Known‑Answer Test (KAT) vector file for the Ascon AEAD algorithm ([example](https://github.com/ascon/ascon-c/blob/main/crypto_aead/asconaead128/LWC_AEAD_KAT_128_128.txt)) and produces a SV ROM. It currently omits Key and Nonce ROMS (they're seemingly consts).

Flag `--sb` swaps the byte order of the data portion (padding unchanged).
