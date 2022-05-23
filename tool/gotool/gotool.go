package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed template/create/*
var templateCreateFs embed.FS

//go:embed template/app/*
var templateAppFs embed.FS

func main() {
	args := os.Args
	if len(args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "create":
		cmdCreate(args)
	case "app":
		cmdApp(args)
	default:
		help()
		return
	}
}

func help() {
	const helpStr = `
Create new project:
	gotool create <xxxx.com/projectname>
Create submodule (run in project folder):
	gotool app <appname>
`
	fmt.Println(strings.TrimSpace(helpStr))
}

// 写入模板文件
func writeTplFileToWr(filepath string, tplContent string, model map[string]string) {
	file, err := os.Create(filepath)
	if err != nil {
		panic(err.Error())
	}
	buffer := bufio.NewWriter(file)
	tpl := template.New("test")
	tpl.Parse(tplContent)
	tpl.Execute(buffer, model)
	buffer.Flush()
	file.Close()
}

func tranverseTplDir(fs embed.FS, dirName string, targetDirName string, model map[string]string) {
	dirs, err := fs.ReadDir(dirName)
	if err != nil {
		panic(err.Error())
	}
	for _, f := range dirs {
		fname := render(f.Name(), model)
		if f.IsDir() {
			os.Mkdir(targetDirName+"/"+fname, os.ModeDir)
			tranverseTplDir(fs, dirName+"/"+f.Name(), targetDirName+"/"+fname, model)
		} else {
			filepathTarget := targetDirName + "/" + strings.TrimSuffix(fname, ".tpl")
			byteData, err := fs.ReadFile(dirName + "/" + f.Name())
			if err != nil {
				panic(err.Error())
			}
			writeTplFileToWr(filepathTarget, string(byteData), model)
		}
	}
}

func render(s string, model map[string]string) string {
	for k, v := range model {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

func cmdCreate(args []string) {
	if len(args) < 3 {
		panic("unknown projectname")
	}
	fullProjectname := args[2]
	projectname := fullProjectname[strings.LastIndex(fullProjectname, "/")+1:]
	fmt.Println("create project " + fullProjectname)

	model := make(map[string]string)
	model["empty"] = ""
	model["projectname"] = projectname
	model["fullprojectname"] = fullProjectname

	// 创建项目文件夹
	os.Mkdir(projectname, os.ModeDir)
	tranverseTplDir(templateCreateFs, "template/create", projectname, model)

	// 初始化go项目
	os.Chdir(projectname)
	fmt.Println("go mod init " + fullProjectname)
	exec.Command("go", "mod", "init", fullProjectname).Run()
	// os.Rename("go.mod", "go.mod.bak")
	// fmt.Println("go mod tidy")
	// exec.Command("go", "mod", "tidy").Run()
	fmt.Println("success")
}

func cmdApp(args []string) {
	if len(args) < 3 {
		panic("unknown appname")
	}
	modstr := ""
	fullProjectname := ""
	appname := ""
	submod := false
	if _, err := os.Stat("go.mod"); err == nil {
		// 子模块，从go.mod中获取项目名，从cmd中获取appname
		submod = true
		modbyte, _ := ioutil.ReadFile("go.mod")
		modstr = strings.TrimSpace(string(modbyte))
		appname = args[2]
		fullProjectname = strings.TrimSpace(strings.TrimPrefix(strings.Split(modstr, "\n")[0], "module ")) + "/app/" + appname
		if strings.LastIndex(appname, "/") > 0 {
			appname = appname[strings.LastIndex(appname, "/")+1:]
			fullProjectname = args[2]
		}
		if _, err := os.Stat("app"); err != nil {
			os.Mkdir("app", os.ModeDir)
		}
	} else {
		// 独立项目，从cmd中获取项目名和appname
		submod = false
		fullProjectname = args[2]
		appname = fullProjectname[strings.LastIndex(fullProjectname, "/")+1:]
	}
	projectname := fullProjectname[strings.LastIndex(fullProjectname, "/")+1:]

	fmt.Printf("fullprojectname: %s\n", fullProjectname)
	fmt.Printf("projectname: %s\n", projectname)
	fmt.Printf("appname: %s\n", appname)

	model := make(map[string]string)
	model["empty"] = ""
	model["appname"] = appname
	model["Appname"] = underscore2camel(appname)
	model["projectname"] = projectname
	model["fullprojectname"] = fullProjectname

	if submod {
		// 创建项目文件夹
		os.Chdir("app")
	}
	os.Mkdir(appname, os.ModeDir)
	tranverseTplDir(templateAppFs, "template/app", appname, model)

	// 初始化app项目
	os.Chdir(appname)
	fmt.Println("go mod init " + fullProjectname)
	exec.Command("go", "mod", "init", fullProjectname).Run()
	fmt.Println("go mod tidy")
	exec.Command("go", "mod", "tidy").Run()
	fmt.Println("success")
}

func underscore2camel(s string) string {
	var r []byte
	flag := true
	for _, c := range s {
		if c >= 'a' && c <= 'z' && flag {
			c0 := c - ('a' - 'A')
			r = append(r, byte(c0))
			flag = false
		} else if c == '_' || c == '-' {
			flag = true
		} else {
			flag = false
			r = append(r, byte(c))
		}
	}
	return string(r)
}
