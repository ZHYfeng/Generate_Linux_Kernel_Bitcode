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

// The path of kernel, e.g., linux
var path = flag.String("path", ".", "the path of kernel")

// "is -save-temps or not"
// two kinds of two to generate bitcode
var IsSaveTemps = flag.Bool("isSaveTemp", false, "use -save-temps or -emit-llvm")

var CC = flag.String("CC", "clang", "Name of CC")
var LD = flag.String("LD", "llvm-link", "Name of LD")

var LLD = flag.String("LLD", "ld.lld", "Name of LD")

// ToolChain of clang and llvm-link
// ToolChain   = "/home/yhao016/data/benchmark/hang/kernel/toolchain/clang-r353983c/bin/"
var ToolChain = flag.String("toolchain", "", "Path of clang and llvm-link")

var FlagCC = FlagAll + FlagCCNoNumber

const (
	PrefixCmd = "cmd_"
	SuffixCmd = ".cmd"
	SuffixCC  = ".o.cmd"

	SuffixLD   = ".a.cmd"
	SuffixLTO  = ".lto.o.cmd"
	SuffixKO   = ".ko.cmd"
	NameScript = "build.sh"

	NameClang = "clang"

	// FlagAll -w disable warning
	// FlagAll -g debug info
	FlagAll = " -w -g"

	// FlagCCNoOptzns disable all optimization
	FlagCCNoOptzns = " -mllvm -disable-llvm-optzns"

	// FlagCCNoNumber add label to basic blocks and variables
	FlagCCNoNumber = " -fno-discard-value-names"

	NameLD    = "llvm-link"
	FlagLD    = " -v "
	FlagOutLD = " -o "

	// arch/x86/kernel/head_64.bc arch/x86/kernel/head64.bc arch/x86/kernel/ebda.bc arch/x86/kernel/platform-quirks.bc
	CmdLinkVmlinux = " -v -o built-in.bc"

	// CmdTools skip the cmd with CmdTools
	CmdTools = "BUILD_STR(s)=$(pound)s"
)

var bitcodes map[string]bool
var linkedBitcodes map[string]bool
var builtinModules map[string]bool

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
				if i > -1 {
					cmd := eachLine[i+3:]
					res = cmd
					break
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
	res = res[strings.Index(res, ""):]
	return res
}

func handleCC(cmd string) string {
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

		// for ";"
		if strings.Count(res, " ; ") == 1 {
			i := strings.Index(res, ";")
			res = res[:i] + "\n"
		}
		res = strings.TrimSpace(res) + "\n"

		// can not compile .S, so just make a empty bitcode file
		if strings.HasSuffix(res, ".S\n") {
			s1 := strings.Split(res, " ")
			s2 := s1[len(s1)-2]
			s3 := strings.Split(s2, ".")
			s4 := s3[0]
			res = "echo \"\" > " + s4 + ".bc" + "\n"
		}
	} else {
		fmt.Println("CC Index not found")
		fmt.Println(cmd)
	}

	res = strings.Replace(res, *CC, filepath.Join(*ToolChain, NameClang), -1)
	res = strings.Replace(res, " -Os ", " -O0 ", -1)
	res = strings.Replace(res, " -O3 ", " -O0 ", -1)
	res = strings.Replace(res, " -O2 ", " -O0 ", -1)
	res = strings.Replace(res, " -fno-var-tracking-assignments ", "  ", -1)
	res = strings.Replace(res, " -fconserve-stack ", "  ", -1)
	res = strings.Replace(res, " -march=armv8-a+crypto ", "  ", -1)
	res = strings.Replace(res, " -mno-fp-ret-in-387 ", "  ", -1)
	res = strings.Replace(res, " -mskip-rax-setup ", "  ", -1)
	res = strings.Replace(res, " -ftrivial-auto-var-init=zero ", "  ", -1)

	return res
}

func handleSuffixCCWithLD(cmd string, path string) string {
	res := ""
	if strings.Index(cmd, "@") > -1 {
		fileName := cmd[strings.Index(cmd, "@")+1 : len(cmd)-1]
		filePath := filepath.Join(path, fileName)
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("handleSuffixCCWithLD file error: ")
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

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		res += filepath.Join(*ToolChain, NameLD)
		res += FlagLD
		res += FlagOutLD
		res += cmd[strings.Index(cmd, FlagOutLD)+len(FlagOutLD) : strings.Index(cmd, "@")]

		for _, s := range text {
			res += s + " "
		}

		res = strings.Replace(res, ".o ", ".bc ", -1)
		res += "\n"

	} else {
		fmt.Println("handleSuffixCCWithLD cmd error: " + cmd)
	}
	return res
}

func handleLD(cmd string) string {

	replace := func(cmd string, i int, length int) string {
		res := ""
		cmd = cmd[i+length:]
		if strings.Count(cmd, ".") > 1 {
			res += filepath.Join(*ToolChain, NameLD)
			res += FlagLD
			res += FlagOutLD
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

		// for multiply cmd or ";" pick the first one
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

func handleLTO(cmd string) string {
	res := ""
	res += filepath.Join(*ToolChain, NameLD)
	res += FlagLD

	cmd = cmd[strings.Index(cmd, FlagOutLD):]
	cmd = strings.Replace(cmd, " --whole-archive ", "", -1)
	cmd = strings.Replace(cmd, ".o", ".bc", -1)

	res += cmd
	return res
}

func handleKO(cmd string) (string, string) {
	res := ""
	res += filepath.Join(*ToolChain, NameLD)
	res += FlagLD
	res += FlagOutLD

	// for multiply cmd or ";" pick the first one
	if strings.Count(cmd, ";") > 1 {
		i := strings.Index(cmd, ";")
		cmd = cmd[:i] + "\n"
	}

	cmd = cmd[strings.Index(cmd, FlagOutLD)+len(FlagOutLD):]
	cmd = strings.Replace(cmd, ".ko", ".ko.bc", -1)
	cmd = strings.Replace(cmd, ".o", ".bc", -1)

	moduleFile := cmd[:strings.Index(cmd, ".ko.bc")+len(".ko.bc")]
	res += cmd

	return res, moduleFile
}

func build(kernelPath string) (string, string) {
	res1 := ""

	var SuffixCCWithLD []string

	err := filepath.Walk(kernelPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixCC) && !strings.HasSuffix(info.Name(), SuffixLTO) {
				cmd := getCmd(path)
				if strings.HasPrefix(cmd, *CC) {
					res2 := handleCC(cmd)
					//res2 = strings.Replace(res2, IncludeOld, IncludeNew, -1)
					res1 += res2
				} else if strings.Index(cmd, *LLD) > -1 {
					SuffixCCWithLD = append(SuffixCCWithLD, cmd)
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

	res3 := ""
	for _, cmd := range SuffixCCWithLD {
		res3 = handleSuffixCCWithLD(cmd, kernelPath) + res3
	}

	res2 := ""
	moduleFiles := ""
	err = filepath.Walk(kernelPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixLD) {
				//for built-in module built-in.a
				cmd := getCmd(path)
				cmd = handleLD(cmd)
				res2 = cmd + res2
				if strings.Index(cmd, FlagOutLD) > -1 {
					cmd = cmd[strings.Index(cmd, FlagOutLD)+len(FlagOutLD):]
					obj := cmd[:strings.Index(cmd, " ")]
					if _, ok := linkedBitcodes[obj]; ok {

					} else {
						builtinModules[obj] = true

					}

					objs := strings.Split(cmd[strings.Index(cmd, " "):len(cmd)-1], " ")
					for _, bc := range objs {
						linkedBitcodes[bc] = true
					}
				}

			} else if strings.HasSuffix(info.Name(), SuffixLTO) {
				//for external module *.lto
				cmd := getCmd(path)
				res2 = handleLTO(cmd) + res2

			} else if strings.HasSuffix(info.Name(), SuffixKO) {
				//for external module *.ko
				cmd, moduleFile := handleKO(getCmd(path))
				res2 = cmd + res2
				moduleFiles = moduleFile + " " + moduleFiles
			}

			return nil
		})

	if err != nil {
		log.Println(err)
	}

	fmt.Println("moduleFiles: ")
	fmt.Println(moduleFiles)

	var res5 string
	for module, _ := range builtinModules {
		res5 += " " + module
	}

	return res1 + res3 + res2 + "\n# external modules: " + moduleFiles + "\n", res5
}

func generateScript(path string, cmd string) {
	res := "#!/bin/bash\n"
	res += cmd

	pathScript := filepath.Join(NameScript)
	_ = os.RemoveAll(pathScript)
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
	_, _ = f.WriteString("\n# path: " + path + "\n")
}

func main() {
	flag.Parse()
	*path, _ = filepath.Abs(*path)

	bitcodes = make(map[string]bool)
	linkedBitcodes = make(map[string]bool)
	builtinModules = make(map[string]bool)

	switch *cmd {
	case "module":
		{
			fmt.Printf("Build module\n")
			res, _ := build(*path)
			generateScript(*path, res)
		}
	case "kernel":
		{
			fmt.Printf("Build kernel and external module\n")
			res, res5 := build(*path)
			res += filepath.Join(*ToolChain, NameLD) + CmdLinkVmlinux + res5 + "\n"
			generateScript(*path, res)
		}
	default:
		fmt.Printf("cmd is invalid\n")
	}
}
