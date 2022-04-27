package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Build one module or whole kernel, e.g., module, kernel
var cmd = flag.String("cmd", "kernel", "Build one module or whole kernel, e.g., module, kernel")

// "The path of kernel, e.g., linux"
var path = "."

// "is -save-temps or not"
// two kinds of two to generate bitcode
var IsSaveTemps = flag.Bool("isSaveTemp", true, "use -save-temps or -emit-llvm")

var CC = flag.String("CC", "clang", "Name of CC")
var LD = flag.String("LD", "llvm-link", "Name of LD")

// ToolChain of clang and llvm-link
// ToolChain   = "/home/yhao016/data/benchmark/hang/kernel/toolchain/clang-r353983c/bin/"
var ToolChain = flag.String("toolchain", "", "Path of clang and llvm-link")

var FlagCC = FlagAll + FlagCCNoNumber

const (
	PrefixCmd  = "cmd_"
	SuffixCmd  = ".cmd"
	SuffixCC   = ".o.cmd"
	SuffixLD   = ".a.cmd"
	SuffixLTO  = ".lto.o.cmd"
	NameScript = "build.sh"

	NameClang = "clang"

	// FlagAll -w disable warning
	// FlagAll -g debug info
	FlagAll = " -w -g"

	// FlagCCNoOptzns disable all optimization
	FlagCCNoOptzns = " -mllvm -disable-llvm-optzns"

	// FlagCCNoNumber add label to basic blocks and variables
	FlagCCNoNumber = " -fno-discard-value-names"

	NameLD = "llvm-link"
	FlagLD = " -v"

	CmdLinkVmlinux = "llvm-link -v -o built-in.bc arch/x86/kernel/head_64.bc arch/x86/kernel/head64.bc arch/x86/kernel/ebda.bc arch/x86/kernel/platform-quirks.bc init/built-in.bc usr/built-in.bc arch/x86/built-in.bc kernel/built-in.bc certs/built-in.bc mm/built-in.bc fs/built-in.bc ipc/built-in.bc security/built-in.bc crypto/built-in.bc block/built-in.bc lib/built-in.bc arch/x86/lib/built-in.bc lib/lib.bc arch/x86/lib/lib.bc drivers/built-in.bc sound/built-in.bc net/built-in.bc virt/built-in.bc arch/x86/pci/built-in.bc arch/x86/power/built-in.bc arch/x86/video/built-in.bc\n"
	// CmdTools skip the cmd with CmdTools
	CmdTools = "-BUILD_STR(s)=$(pound)s"
)

func getCmd(cmdFilePath string) string {
	res := ""
	if _, err := os.Stat(cmdFilePath); os.IsNotExist(err) {
		fmt.Printf(cmdFilePath + " does not exist\n")
	} else {
		file, err := os.Open(cmdFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		var text []string
		for scanner.Scan() {
			text = append(text, scanner.Text())
		}
		for _, eachLine := range text {
			if strings.HasPrefix(eachLine, PrefixCmd) {
				i := strings.Index(eachLine, ":=")
				// fmt.Println("Index: ", i)
				if i > -1 {
					cmd := eachLine[i+3:]
					res = cmd
				} else {
					fmt.Println("Cmd Index not found")
					fmt.Println(eachLine)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	res += "\n"
	return res
}

func replaceCC(cmd string) string {
	res := ""
	if i := strings.Index(cmd, " -c "); i > -1 {

		if j := strings.Index(cmd, CmdTools); j > -1 {
			return res
		}

		res += cmd[:i]
		res += FlagCC
		if *IsSaveTemps {
			res += " -save-temps=obj"
		} else {
			res += " -emit-llvm"
		}
		res += cmd[i:]

		// replace .o to .bc
		if *IsSaveTemps {

		} else {
			res = strings.Replace(res, ".o ", ".bc ", -1)
		}

		// can not compile .S, so just make a empty bitcode file
		if strings.HasSuffix(cmd, ".S\n") {
			s1 := strings.Split(cmd, " ")
			s2 := s1[len(s1)-2]
			s3 := strings.Split(s2, ".")
			s4 := s3[0]

			res += "\n"
			res = "echo \"\" > " + s4 + ".bc" + "\n"
		}
	} else {
		fmt.Println("CC Index not found")
		fmt.Println(cmd)
	}
	// for ";"
	if strings.Count(res, ";") > 1 {
		i := strings.Index(res, ";")
		res = res[:i] + "\n"
	}
	return res
}

func replaceLD(cmd string) string {

	replace := func(cmd string, i int, length int) string {
		res := ""
		cmd = cmd[i+length:]
		if strings.Count(cmd, ".") > 1 {
			res += NameLD
			res += FlagLD
			res += " -o "
			res += cmd
			if strings.Contains(res, "drivers/of/unittest-data/built-in.o") {
				res = ""
			}
			res = strings.Replace(res, ".o", ".bc", -1)
		} else {
			res = "echo \"\" > " + cmd
			res = strings.Replace(res, ".o", ".bc ", -1)
		}
		res = strings.Replace(res, ".a ", ".bc ", -1)
		res = strings.Replace(res, ".a\n", ".bc\n", -1)
		// for this drivers/misc/lkdtm/rodata.bc
		res = strings.Replace(res, "rodata_objcopy.bc", "rodata.bc", -1)
		res = strings.Replace(res, " drivers/of/unittest-data/built-in.bc", "", -1)

		// for ";"
		if strings.Count(res, ";") > 1 {
			i := strings.Index(res, ";")
			res = res[:i] + "\n"
		}
		return res
	}

	res := ""
	// fmt.Println("Index: ", i)
	if i := strings.Index(cmd, " rcSTPD "); i > -1 {
		res = replace(cmd, i, len(" rcSTPD "))
	} else if i := strings.Index(cmd, " cDPrST "); i > -1 {
		res = replace(cmd, i, len(" cDPrST "))
	} else if i := strings.Index(cmd, " cDPrsT "); i > -1 {
		res = replace(cmd, i, len(" cDPrsT "))
	} else if i := strings.Index(cmd, " rcsD "); i > -1 {
		res = replace(cmd, i, len(" rcsD "))
	} else if i := strings.Index(cmd, *LD); i > -1 {
		res = replace(cmd, i, len(*LD))
	} else {
		fmt.Println("LD Index not found")
		fmt.Println(cmd)
	}

	return res
}

func get_linked_target(cmd string) string {
	res := ""
	if strings.Contains(cmd, "llvm-link -v -o") {
		res = cmd[len("llvm-link -v -o") : strings.Index(cmd, ".bc")+3]
	}
	return res
}

func build(moduleDirPath string) string {
	res1 := ""
	err := filepath.Walk(moduleDirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixCC) {
				cmd := getCmd(path)
				if strings.HasPrefix(cmd, *CC) {
					res2 := replaceCC(cmd)
					res2 = strings.Replace(res2, *CC, NameClang, -1)
					res2 = strings.Replace(res2, " -Os ", " -O0 ", -1)
					res2 = strings.Replace(res2, " -O3 ", " -O0 ", -1)
					res2 = strings.Replace(res2, " -O2 ", " -O0 ", -1)
					res2 = strings.Replace(res2, " -fno-var-tracking-assignments ", "  ", -1)
					res2 = strings.Replace(res2, " -fconserve-stack ", "  ", -1)
					res2 = strings.Replace(res2, " -march=armv8-a+crypto ", "  ", -1)
					//res2 = strings.Replace(res2, IncludeOld, IncludeNew, -1)
					res1 += res2
				} else {
					fmt.Println(*CC + " not found")
					fmt.Println(path)
					fmt.Println(cmd)
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	res2 := ""
	module_file := ""
	err = filepath.Walk(moduleDirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixLD) {
				cmd := getCmd(path)
				res2 = replaceLD(cmd) + res2
			}
			// for kernel module (*.ko, *.lto)
			if strings.HasSuffix(info.Name(), SuffixCC) {
				cmd := getCmd(path)
				res2 = replaceLD(cmd) + res2
			}
			if strings.HasSuffix(info.Name(), SuffixLTO) {
				cmd := getCmd(path)
				cmd = cmd[strings.Index(cmd, "--whole-archive")+len("--whole-archive") : len(cmd)-1]
				cmd = strings.Replace(cmd, ".o", ".bc", -1)
				module_file = cmd + module_file
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	fmt.Println("module_file")
	fmt.Println(module_file)

	return res1 + res2
}

func generateScript(path string, cmd string) {
	res := "#!/bin/bash\n"
	res += cmd

	pathScript := filepath.Join(path, NameScript)
	_ = os.Remove(pathScript)
	fmt.Printf("script path : bash %s\n", pathScript)
	f, err := os.OpenFile(pathScript, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	_, _ = f.WriteString(res)
}

func main() {
	flag.Parse()

	switch *cmd {
	case "module":
		{
			fmt.Printf("Build one module\n")
			res := build(path)
			generateScript(path, res)
		}
	case "kernel":
		{
			fmt.Printf("Build whole kernel\n")
			res := build(path)
			res += CmdLinkVmlinux
			generateScript(path, res)
		}
	default:
		fmt.Printf("cmd is invalid\n")
	}
}
