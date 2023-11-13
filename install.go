package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var basePath string
var newOldBasePath string

func installUpdate() error {
	fmt.Println("Installer started!")
	currentOS := runtime.GOOS
	exeFile, err := os.Executable()
	if err != nil {
		return err
	}
	currentDir := filepath.Join(exeFile, "..")

	if currentOS == "windows" {
		basePath = filepath.Join(currentDir, "base.exe")
		newOldBasePath = filepath.Join(currentDir, "old-base.exe")
	} else {
		basePath = filepath.Join(currentDir, "base")
		newOldBasePath = filepath.Join(currentDir, "old-base")
	}

	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Errorf("panic: %v", r)
			fmt.Println(errMsg)
			fmt.Println("FAILED TO INSTALL, REVERTING CHANGES")
			err = os.Rename(newOldBasePath, basePath)
			if err != nil {
				log.Println("Unable to recover executable")
			}
		}
	}()

	//Rename to old-base
	err = os.Rename(basePath, newOldBasePath)
	if err != nil {
		log.Panic("Unable to rename executable")
		return err
	}
	log.Println("Renamed running base to old-base")

	//Download the base
	updateApiKeyEnv := os.Getenv("updateApiKey")
	updateApiUrl := os.Getenv("updateApiUrl")
	err = downloadTheFile(basePath, updateApiUrl+updateApiKeyEnv)
	if err != nil {
		log.Panic("Failed to download new exe zip", err)
		return err
	}
	log.Println("New base downloaded.")

	err = os.Chmod(basePath, 0755)
	if err != nil {
		fmt.Println("Error give permision (EXE:BASE):", err)
		return err
	}
	log.Println("Chmod permisions given to new base")

	//Start the exe
	defer exec.Command("reboot").Start()
	exec.Command("reboot").Start()

	return nil
}

//The plan:
//Rename it self to old-base
//download the zip
//extract it to current dir
//run the exe check if is windows or not
//when running exe run with args newinstall to run custom script
//Newinstall should not imediatly start the app but kill old-base.exe then start pb

func downloadTheFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func runTheExecutable(executablePath string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var err error

	//currentOS := runtime.GOOS
	//if currentOS != "windows" {
	//	err = cmd.Start()
	//} else {
	//	cmd.Run()
	//}
	cmd.Start()
	return cmd, err
}

func KillTheOldExe() {
	// Get the operating system
	currentOS := runtime.GOOS
	exeFile, err := os.Executable()
	if err != nil {
		return
	}
	currentDir := filepath.Join(exeFile, "..")

	if currentOS == "windows" {
		basePath = filepath.Join(currentDir, "base.exe")
		newOldBasePath = filepath.Join(currentDir, "old-base.exe")
	} else {
		basePath = filepath.Join(currentDir, "base")
		newOldBasePath = filepath.Join(currentDir, "old-base")
	}

	// Check if it's Windows
	if currentOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", "old-base.exe")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Start()
		if err != nil {
			log.Println("Error terminating process (EXE:INSTALLER):", err)
			log.Println("^ Propbs couldn't find it because it's not running anyway")
		}
		log.Println("old-base killed")
	} else {
		// Run the pkill command to terminate the process by name
		cmd := exec.Command("pkill", "old-base")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Start()
		if err != nil {
			log.Println("Error terminating process (EXE:INSTALLER):", err)
			log.Println("^ Propbs couldn't find it because it's not running anyway")
		}
		os.Remove(newOldBasePath)
		log.Println("old-base killed")

	}
}
