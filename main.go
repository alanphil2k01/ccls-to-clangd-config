package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	board string
	ide   string
	path  string
)

func checkPath(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func runPioInit() {
	if !commandExists("platformio") {
		panic("platformio command not found")
	}
	pio := exec.Command("platformio", "init", "-b", board, "--ide", ide)
	if err := pio.Run(); err != nil {
		panic("Error exectuing platformio init" + err.Error())
	}
}

func pwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("Error cannot get current working directory: " + err.Error())
	}
	return dir
}

var indent = ""

func writeHeaders(clangd *bufio.Writer) {
	clangd.WriteString(indent + "CompileFlags:\n")
	indent += " "
	clangd.WriteString(indent + "Add:\n")
	indent += " "
}

func writeLine(clangd *bufio.Writer, str []byte) {
	clangd.WriteString(indent + "- " + string(str))
}

func convertCCLS() error {
	ccls, err := os.Open(".ccls")
	if err != nil {
		return err
	}
	clangd, err := os.Create(".clangd")
	if err != nil {
		return err
	}

	defer func() {
		if err := ccls.Close(); err != nil {
			panic(err)
		}
		if err := clangd.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(ccls)
	writer := bufio.NewWriter(clangd)
	writeHeaders(writer)

	var buf []byte
	for {
		buf, err = reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if len(buf) != 0 && buf[0] == '-' {
			writeLine(writer, buf)
		}

	}

	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.StringVar(&path, "d", pwd(), "path of project. $PWD if not  specified")
	flag.StringVar(&ide, "ide", "vim", "ide")
	flag.StringVar(&board, "board", "uno", "board")
	flag.Parse()

	if !checkPath(path) {
		panic(errors.New("Invalid directory " + path))
	}

	if !checkPath(path + "/platformio.ini") {
		fmt.Println("Not a platformio project. Running init anyway")
	}

	if err := os.Chdir(path); err != nil {
		panic("Error in changing directory: " + err.Error())
	}

	fmt.Println("Running platformio init")
	runPioInit()
	if !checkPath(path + "/.ccls") {
		fmt.Println("Retrying platformio init with ide as vim")
		ide = "vim"
		runPioInit()
	}

	fmt.Println("converting .ccls to .clangd")
	if err := convertCCLS(); err != nil {
		panic("Cannot convert .ccls to .clangd" + err.Error())
	}
	fmt.Println("Done")
}
