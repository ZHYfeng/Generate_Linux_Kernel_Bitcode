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

const (
	PrefixCmd  = "cmd_"
	SuffixCmd  = ".cmd"
	SuffixCC   = ".o.cmd"
	SuffixLD   = ".a.cmd"
	NameScript = "build.sh"

	NameCC = "clang"
	// FlagCC = " -save-temps=obj -w -mllvm -disable-llvm-optzns"
	FlagCC = " -save-temps=obj -w"
	NameLD = "llvm-link"
	// Path   = "/home/yhao016/data/benchmark/hang/kernel/toolchain/clang-r353983c/bin/"
	Path = ""
	// path of clang and llvm-link
)

var CC = filepath.Join(Path, NameCC)
var LD = filepath.Join(Path, NameLD)

func getCmd(cmdFilePath string) string {
	res := ""
	if _, err := os.Stat(cmdFilePath); os.IsNotExist(err) {
		fmt.Printf(cmdFilePath + " does not exist\n")
	} else {
		file, err := os.Open(cmdFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

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
					fmt.Println(eachLine)
					fmt.Println("Cmd Index not found")
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

func replaceCC(cmd string, addFlag bool) string {
	res := ""
	if addFlag {
		i := strings.Index(cmd, " -c ")
		// fmt.Println("Index: ", i)
		if i > -1 {
			res += cmd[:i]
			res += FlagCC
			res += cmd[i:]
		} else {
			fmt.Println(cmd)
			fmt.Println("CC Index not found")
		}
	}
	return res
}

func replaceLD(cmd string) string {

	res := ""
	// fmt.Println("Index: ", i)
	if i := strings.Index(cmd, " rcSTPD"); i > -1 {
		cmd = cmd[i+8:]
		if strings.Count(cmd, ".") > 1 {
			res += LD
			res += " -o "
			res += cmd
			res = strings.Replace(res, ".o", ".bc", -1)
		} else {
			res = "echo \"\" > " + cmd
		}
		res = strings.Replace(res, "built-in.a", "built-in.bc", -1)
	} else if i := strings.Index(cmd, " cDPrST"); i > -1 {
		cmd = cmd[i+8:]
		if strings.Count(cmd, ".") > 1 {
			res += LD
			res += " -o "
			res += cmd
			res = strings.Replace(res, ".o", ".bc", -1)
		} else {
			res = "echo \"\" > " + cmd
		}
		res = strings.Replace(res, "built-in.a", "built-in.bc", -1)
	} else if i := strings.Index(cmd, " cDPrsT"); i > -1 {
		cmd = cmd[i+8:]
		if strings.Count(cmd, ".") > 1 {
			res += LD
			res += " -o "
			res += cmd
			res = strings.Replace(res, ".o", ".bc", -1)
		} else {
			res = "echo \"\" > " + cmd
		}
		res = strings.Replace(res, "built-in.a", "built-in.bc", -1)
	} else {
		fmt.Println(cmd)
		fmt.Println("LD Index not found")
	}

	return res
}

func buildModule(moduleDirPath string) string {
	res1 := ""
	err := filepath.Walk(moduleDirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixCC) {
				cmd := getCmd(path)
				if strings.HasPrefix(cmd, NameCC) {
					res1 += replaceCC(cmd, true)
				} else {
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
	err = filepath.Walk(moduleDirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), SuffixLD) {
				cmd := getCmd(path)
				res2 = replaceLD(cmd) + res2
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
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
	defer f.Close()

	_, _ = f.WriteString(res)
}

var cmd = flag.String("cmd", "module", "Build one module or whole kernel, e.g., module, kernel")
var path = flag.String("path", ".", "The path of data, e.g., module, make.log.")

func main() {
	flag.Parse()
	switch *cmd {
	case "module":
		{
			fmt.Printf("Build one module\n")
			res := buildModule(*path)
			generateScript(*path, res)
		}
	case "kernel":
		{
			fmt.Printf("Build whole kernel with make.log\n")
		}
	default:
		fmt.Printf("cmd is invalid\n")
	}
}
