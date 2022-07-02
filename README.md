[TOC]

# Requirement
- build kernel using clang into binary successfully
- golang
- The kernel module are enabled in Linux kernel configuration

# Build Linux kernel into LLVM Bitcode
```shell
go run path/of/code/02-replace_cmd_log/GenKernelBitcode.go --help
```

## Generate script
```shell
cd path/Linux/kernel
go run path/of/code/02-replace_cmd_log/GenKernelBitcode.go -path=./path/kernel/module
```

## Get LLVM Bitcode
```shell
bash build.sh
```
`built-in.bc` is in `./path/kernel/module/`

external modules are in the end of build.sh