package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	zipURL    = "https://github.com/jfmow/pocketbase/raw/main/base.zip" // Replace with your actual URL
	zipName   = "base.zip"
	targetDir = "update"
)

func updater() {
	// Download the zip file
	err := downloadFile(zipName, zipURL)
	if err != nil {
		fmt.Println("Error downloading zip file:", err)
		return
	}
	defer os.Remove(zipName)

	// Unzip the file
	err = unzip(zipName, targetDir)
	if err != nil {
		fmt.Println("Error unzipping file:", err)
		return
	}

	// Run the updated executable

	currentOS := runtime.GOOS
	executableName := "installer"
	if currentOS == "windows" {
		executableName = executableName + ".exe"
	}
	fmt.Println(executableName)

	err = os.Chmod(filepath.Join(targetDir, executableName), 0755)
	if err != nil {
		fmt.Println("Error give permision:", err)
		return
	}
	_, err = runExecutable(filepath.Join(targetDir, executableName))
	if err != nil {
		fmt.Println("Error running executable:", err)
		return
	}

	fmt.Println("Program executed successfully.")
	fmt.Println("Shuting down...")
	os.Exit(0)
}

func downloadFile(filepath string, url string) error {
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

func replaceExecutable(newExecutable string) error {
	currentOS := runtime.GOOS
	executableName := "installer"
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
	var err error

	currentOS := runtime.GOOS
	if currentOS != "windows" {
		err = cmd.Start()
	} else {
		cmd.Run()
	}
	return cmd, err
}
