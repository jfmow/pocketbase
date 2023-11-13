package main

import (
	"archive/zip"
	"fmt"
	"io"
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
	err := unzip(zipName, targetDir)
	if err != nil {
		fmt.Println("Error unzipping file (EXE:INSTALLER):", err)
		return
	}
	currentOS := runtime.GOOS
	executableName := "base"
	if currentOS == "windows" {
		executableName = executableName + ".exe"
	}

	exeDir, err := os.Executable()
	if err != nil {
		fmt.Println("Error replacing find executable (EXE:INSTALLER):", err)
		return
	}
	fmt.Println("Found exe")
	err = replaceExecutable(filepath.Join(exeDir, "..", "update", executableName))
	if err != nil {
		fmt.Println("Error replacing executable (EXE:INSTALLER):", err)
		return
	}
	fmt.Println("Swapped exe's")

	// Run the updated executable
	log.Println(filepath.Join(exeDir, "..", executableName))
	err = os.Chmod(filepath.Join(exeDir, "..", executableName), 0755)
	if err != nil {
		fmt.Println("Error give permision 2 (EXE:INSTALLER):", err)
		return
	}
	fmt.Println("Permisions given")
	_, err = runExecutable(filepath.Join(exeDir, "..", executableName), "serve")
	if err != nil {
		fmt.Println("Error running executable (EXE:INSTALLER):", err)
		return
	}

	fmt.Println("New base executed successfully.\nBye...")
	os.Remove(filepath.Join(exeDir, "..", zipName))
	os.Exit(0)
}

func replaceExecutable(newExecutable string) error {

	currentOS := runtime.GOOS
	executableName := "base"
	if currentOS == "windows" {
		executableName = executableName + ".exe"
	}
	exeDir, err := os.Executable()
	if err != nil {
		return err
	}
	currentExecutable := filepath.Join(exeDir, "..", executableName)
	err = os.Rename(newExecutable, currentExecutable)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runExecutable(executablePath string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
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
			fmt.Println("Error terminating process (EXE:INSTALLER):", err)
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
			fmt.Println("Error terminating process (EXE:INSTALLER):", err)
		}
		fmt.Println("Process killed")
	}
}
