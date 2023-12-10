package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed template/create/*
var templateCreateFs embed.FS

func main() {
	args := os.Args
	if len(args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "create":
		cmdCreate(args)
	default:
		help()
		return
	}
}

func help() {
	const helpStr = `
Create new project:
	gotool create <xxxx.com/projectname>
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
		s = strings.ReplaceAll(s, "{{."+k+"}}", v)
	}
	return s
}

func cmdCreate(args []string) {
	if len(args) < 3 {
		panic("unknown appname")
	}
	// 独立项目，从cmd中获取项目名和appname
	// submod = false
	fullProjectname := args[2]
	appname := fullProjectname[strings.LastIndex(fullProjectname, "/")+1:]
	projectname := fullProjectname[strings.LastIndex(fullProjectname, "/")+1:]

	fmt.Printf("fullprojectname: %s\n", fullProjectname)
	fmt.Printf("projectname:     %s\n", projectname)
	fmt.Printf("appname:         %s\n", appname)

	model := make(map[string]string)
	model["empty"] = ""
	model["appname"] = appname
	model["Appname"] = underscore2camel(appname)
	model["appname_"] = strings.ReplaceAll(appname, "-", "_")
	model["projectname"] = projectname
	model["fullprojectname"] = fullProjectname

	os.Mkdir(appname, os.ModeDir)
	tranverseTplDir(templateCreateFs, "template/create", appname, model)

	// 初始化app项目
	os.Chdir(appname)
	fmt.Println("[exec] go mod init " + fullProjectname)
	exec.Command("go", "mod", "init", fullProjectname).Run()
	fmt.Println("[exec] go mod tidy")
	exec.Command("go", "mod", "tidy").Run()

	// 加入workspace
	os.Chdir("../")
	if _, err := os.Stat("go.work"); err == nil {
		fmt.Println("[exec] go work use ./" + appname)
		exec.Command("go", "work", "use", "./"+appname).Run()
	}
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
