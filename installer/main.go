package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	zipName   = "base.zip"
	targetDir = "update"
)

func main() {
	KillOldExe()
	currentOS := runtime.GOOS
	executableName := "base"
	if currentOS == "windows" {
		executableName = executableName + ".exe"
	}

	exeDir, err := os.Executable()
	if err != nil {
		fmt.Println("Error replacing find executable:", err)
		return
	}
	err = replaceExecutable(filepath.Join(exeDir, "..", executableName))
	if err != nil {
		fmt.Println("Error replacing executable:", err)
		return
	}

	// Run the updated executable
	log.Println(filepath.Join(exeDir, "..", "..", executableName), filepath.Join("..", executableName))
	err = os.Chmod(filepath.Join(exeDir, "..", "..", executableName), 0755)
	if err != nil {
		fmt.Println("Error give permision 2:", err)
		return
	}
	_, err = runExecutable(filepath.Join(exeDir, "..", "..", executableName), "serve")
	if err != nil {
		fmt.Println("Error running executable:", err)
		return
	}

	fmt.Println("Program executed successfully.")
	fmt.Println("Success")
	os.Exit(0)
}

func replaceExecutable(newExecutable string) error {

	currentOS := runtime.GOOS
	executableName := "base"
	if currentOS == "windows" {
		executableName = executableName + ".exe"
	}
	exeDir, err := os.Executable()
	currentExecutable := filepath.Join(exeDir, "..", "..", executableName)
	err = os.Rename(newExecutable, currentExecutable)
	return err
}

func runExecutable(executablePath string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	return cmd, err
}

func KillOldExe() {
	executableName := "base"

	// Get the operating system
	currentOS := runtime.GOOS

	// Check if it's Windows
	if currentOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", executableName+".exe")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error terminating process:", err)
		}
		fmt.Println("Process killed")
	} else {
		// Run the pkill command to terminate the process by name
		cmd := exec.Command("pkill", executableName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error terminating process:", err)
		}
		fmt.Println("Process killed")
	}
}
