[TOC]

# Requirement
- build kernel using clang into binary successfully
- golang
- The kernel module are enabled in Linux kernel configuration

# Build one kernel module into LLVM Bitcode

## Set path

- set `Path` of `llvm-link` in `02-replace_cmd_log/buildLLVMBitcode.go`

## Generate script
```
cd path/Linux/kernel
go run path/of/code/02-replace_cmd_log/buildLLVMBitcode.go -path=./path/kernel/module
```

## Get LLVM Bitcode
```
bash ./path/kernel/module/build.sh
```
`built-in.bc` is in `./path/kernel/module/`

# Build Linux kernel into LLVM Bitcode (Doing)


```shell
make LLVM=1 V=1 defconfig
make LLVM=1 V=1 -j64
go run ~/data/git/02-replace_cmd_log/buildLLVMBitcode.go -cmd=kernel -path=.
bash build.sh
```
