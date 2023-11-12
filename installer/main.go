package main

import (
	"fmt"
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
	err := replaceExecutable(filepath.Join(targetDir, executableName))
	if err != nil {
		fmt.Println("Error replacing executable:", err)
		return
	}

	// Run the updated executable
	_, err = runExecutable(filepath.Join("..", executableName), "serve")
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
	currentExecutable := filepath.Join("..", executableName)
	err := os.Rename(newExecutable, currentExecutable)
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
