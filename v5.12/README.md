# Requirement
- build kernel using clang into binary successfully
- golang
- The related kernel module are enabled in Linux kernel configuration

# Build Linux kernel into LLVM Bitcode

```shell
go run path/of/v5.12/KernelBitcode.go --help
```


1. Build Kernel with clang
```shell
make LLVM=1 -j16
```
2. Generate script to build kernel into LLVM Bitcode
```shell
go run path/of/v5.12/KernelBitcode.go
```
3. Get LLVM Bitcode
```shell
bash build.sh
```
`built-in.bc` is in each directory. 
All external modules are in the end of build.sh.
Now you can perform analysis on those LLVM bitcode files.