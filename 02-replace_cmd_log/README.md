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
go run ~/data/git/2019-Build_Linux_Kernel_Into_LLVM_Bitcode/02-replace_cmd_log/buildLLVMBitcode.go -cmd=kernel -path=.
bash build.sh
```

# Notice

```shell
llvm-link -v -o drivers/misc/lkdtm/built-in.bc drivers/misc/lkdtm/core.bc drivers/misc/lkdtm/bugs.bc drivers/misc/lkdtm/heap.bc drivers/misc/lkdtm/perms.bc drivers/misc/lkdtm/refcount.bc drivers/misc/lkdtm/rodata.bc drivers/misc/lkdtm/usercopy.bc drivers/misc/lkdtm/stackleak.bc drivers/misc/lkdtm/cfi.bc drivers/misc/lkdtm/fortify.bc
```
