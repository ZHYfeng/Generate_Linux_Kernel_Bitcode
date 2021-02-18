[TOC]

# Requirement
- build kernel using clang into binary successfully
- golang
- enabled in Linux kernel configuration

# Build one kernel module into LLVM Bitcode

## Set path

- set `Path` of `llvm-link` in `02-way-replace_cmd_log/buildLLVMBitcode.go`

## Generate script
```
cd path/Linux/kernel
go run ~/data/git/Build-Linux-Kernel-Using-Clang/02-way-replace_cmd_log/buildLLVMBitcode.go -path=./path/kernel/module
```

## Get LLVM Bitcode
```
bash ./path/kernel/module/build.sh
```
`built-in.bc` is in `./path/kernel/module/`

# Build Linux kernel into LLVM Bitcode (Doing)

## Get makefile log
```shell
make V=1 1>make.log 2>&1
```
> must only one process
