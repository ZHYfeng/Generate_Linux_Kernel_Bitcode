[TOC]

# Clone the Linux source tree
```shell
git clone git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux-stable.git
cd linux-stable
git reset --hard v4.16
```
> After linux kernel v4.16, it needs compiler support asm-goto and clang do not.

# Configure the kernel
O0 can not be used.

# Build the kernel
```shell
cd linux-stable
make SHELL=/bin/bash V=1 2>&1 | tee make.log
```
# Link the bc file
```shell
llvm-link -o built-in.bc arch/x86/kernel/head64.bc arch/x86/kernel/ebda.bc arch/x86/kernel/platform-quirks.bc init/built-in.bc usr/built-in.bc arch/x86/built-in.bc kernel/built-in.bc certs/built-in.bc mm/built-in.bc fs/built-in.bc ipc/built-in.bc security/built-in.bc crypto/built-in.bc block/built-in.bc lib/built-in.bc arch/x86/lib/built-in.bc drivers/built-in.bc sound/built-in.bc firmware/built-in.bc arch/x86/pci/built-in.bc arch/x86/power/built-in.bc arch/x86/video/built-in.bc net/built-in.bc virt/built-in.bc
```
# Get file used for analysis
```
llvm-dis built-in.bc
cat `find -name "*.s"` >> built-in.s
```
