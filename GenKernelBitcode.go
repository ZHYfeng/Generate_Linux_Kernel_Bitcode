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

// IsSaveTemps : use -save-temps or emit-llvm to generate LLVM Bitcode
// two kinds of two to generate bitcode
var IsSaveTemps = flag.Bool("isSaveTemp", false, "use -save-temps or -emit-llvm")

// tools used in build kernel
var CC = flag.String("CC", "clang", "Name of CC")
var LD = flag.String("LD", "llvm-link", "Name of LD")
var AR = flag.String("AR", "llvm-ar", "Name of AR")
var LLD = flag.String("LLD", "ld.lld", "Name of LD")
var OBJCOPY = flag.String("OBJCOPY", "llvm-objcopy", "Name of OBJCOPY")
var STRIP = flag.String("STRIP", "llvm-strip", "Name of STRIP")

// ToolChain of clang and llvm-link
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

	// FlagAll : -w disable warning
	// FlagAll : -g debug info
	FlagAll = " -w -g"

	// FlagCCNoOptzns disable all optimization
	FlagCCNoOptzns = " -mllvm -disable-llvm-optzns"

	// FlagCCNoNumber add label to basic blocks and variables
	FlagCCNoNumber = " -fno-discard-value-names"

	NameLD    = "llvm-link"
	FlagLD    = " -v "
	FlagOutLD = " -o "

	CmdLinkVmlinux = " -v -o built-in.bc"

	// CmdTools skip the cmd with CmdTools
	CmdTools = "BUILD_STR(s)=$(pound)s"
)

var bitcodes map[string]bool
var linkedBitcodes map[string]bool
var builtinModules map[string]bool

// get cmd from *.cmd file
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

// for CC cmd, use " -save-temps=obj" or " -emit-llvm" to generate llvm bitcode
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

		// for multiply ";"
		if strings.Count(res, " ; ") == 1 {
			i := strings.Index(res, ";")
			res = res[:i] + "\n"
		}
		res = strings.TrimSpace(res) + "\n"

		// can not compile .S, so just make a empty bitcode file
		if strings.HasSuffix(res, ".S\n") {
			s1 := strings.Split(res, " ")
			s2 := s1[len(s1)-2]
			s4 := strings.Replace(s2, ".o ", ".bc ", -1)
			res = "echo \"\" > " + s4 + "\n"
		}
	} else {
		fmt.Println("CC Index not found")
		fmt.Println(cmd)
	}
	res = " " + res
	// use -O0 install of other optimization
	res = strings.Replace(res, " "+*CC+" ", " "+filepath.Join(*ToolChain, NameClang)+" ", -1)
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

// handler LD cmd in *.o.cmd
// @file_name in *.o.cmd includes the related file
// need to get the name from that file
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

	} else if strings.HasPrefix(cmd, *LLD) {
		res += filepath.Join(*ToolChain, NameLD)
		res += FlagLD
		res += FlagOutLD

		cmd = cmd[:len(cmd)-1]
		s1 := strings.Split(cmd, " ")
		obj := ""
		for _, s := range s1 {
			if strings.HasSuffix(s, ".o") {
				obj = " " + strings.Replace(s, ".o", ".bc", -1) + obj
			}
		}
		res += obj
		res += "\n"
	} else {
		fmt.Println("handleSuffixCCWithLD cmd error: " + cmd)
	}
	return res
}

// handle llvm-objcopy cmd
func handleOBJCOPY(cmd string) string {
	res := filepath.Join(*ToolChain, NameLD) + FlagLD + FlagOutLD
	cmd = cmd[:len(cmd)-1]
	s1 := strings.Split(cmd, " ")
	obj := ""
	for _, s := range s1 {
		if strings.HasSuffix(s, ".o") {
			obj = " " + strings.Replace(s, ".o", ".bc", -1) + obj
		}
	}
	res += obj
	res += "\n"
	return res
}

// handle llvm-strip cmd
func handleSTRIP(cmd string) string {
	res := filepath.Join(*ToolChain, NameLD) + FlagLD + FlagOutLD
	s1 := strings.Split(cmd, ";")
	cmd = s1[0]
	s1 = strings.Split(cmd, " ")
	for _, s := range s1 {
		if strings.HasSuffix(s, ".o") {
			res = res + " " + strings.Replace(s, ".o", ".bc", -1)
		}
	}
	res += "\n"
	return res
}

// use llvm-link to link all bitcode
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

// handler *.lto for external modules
func handleLTO(cmd string) string {
	res := ""
	res += filepath.Join(*ToolChain, NameLD)
	res += FlagLD
	res += FlagOutLD

	cmd = cmd[strings.Index(cmd, FlagOutLD) : len(cmd)-1]
	objs := strings.Split(cmd, " ")
	output := false
	for _, obj := range objs {
		if obj == "-o" {
			output = true
		} else if output && obj != "" {
			res += strings.Replace(obj, ".o", ".bc", -1)
			output = false
		} else if strings.HasSuffix(obj, ".o") {
			res += " " + strings.Replace(obj, ".o", ".bc", -1)
		}
	}
	res += "\n"
	return res
}

// handle .ko for external modules
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

// find all *.cmd file and handle the cmd in them
func build(kernelPath string) (string, string) {
	cmdCC := ""
	cmdLDInCC := ""

	err := filepath.Walk(kernelPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			//  handle, all *.o.cmd files.
			//  do not include  *.lto.o.cmd files
			if strings.HasSuffix(info.Name(), SuffixCC) && !strings.HasSuffix(info.Name(), SuffixLTO) {
				//  get cmd from the file
				cmd := getCmd(path)
				if strings.HasPrefix(cmd, *CC) {
					cmd := handleCC(cmd)
					cmdCC += cmd
				} else if strings.Index(cmd, *AR) > -1 {
					cmd = handleLD(cmd)
					cmdLDInCC = cmd + cmdLDInCC
				} else if strings.Index(cmd, *LLD) > -1 {
					cmd = handleSuffixCCWithLD(cmd, kernelPath)
					cmdLDInCC = cmd + cmdLDInCC
				} else if strings.HasPrefix(cmd, *OBJCOPY) {
					cmd = handleOBJCOPY(cmd)
					cmdLDInCC = cmd + cmdLDInCC
				} else if strings.HasPrefix(cmd, *STRIP) {
					cmd = handleSTRIP(cmd)
					cmdLDInCC = cmd + cmdLDInCC
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

	cmdLink := ""
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
				cmdLink = cmd + cmdLink
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
				cmdLink = handleLTO(cmd) + cmdLink

			} else if strings.HasSuffix(info.Name(), SuffixKO) {
				//for external module *.ko
				cmd, moduleFile := handleKO(getCmd(path))
				cmdLink = cmd + cmdLink
				moduleFiles = moduleFile + " " + moduleFiles
			}

			return nil
		})

	if err != nil {
		log.Println(err)
	}

	fmt.Println("moduleFiles: ")
	fmt.Println(moduleFiles)

	var resFinal string
	for module, _ := range builtinModules {
		resFinal += " " + module
	}

	return cmdCC + cmdLDInCC + cmdLink + "\n# external modules: " + moduleFiles + "\n", resFinal
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

// collect compile cmd from *.cmd files in kernel
// get cmd to generate llvm bitcode
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
