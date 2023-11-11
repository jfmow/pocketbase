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

const zipURL = "https://raw.githubusercontent.com/jfmow/pocketbase/master/base.zip" // Replace with your actual URL

func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func unzipFile(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		filePath := filepath.Join(destPath, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
		} else {
			os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
		rc.Close()
	}
	return nil
}

func restart() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	cmd := exec.Command(execPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error restarting process:", err)
		return
	}

	// Gracefully shut down the old process
	os.Exit(0)
}

func updater() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	zipPath := filepath.Join(os.TempDir(), "update.zip")
	err = downloadFile(zipURL, zipPath)
	if err != nil {
		fmt.Println("Error downloading zip file:", err)
		return
	}

	tempDir := filepath.Join(os.TempDir(), "update")
	os.RemoveAll(tempDir)
	os.Mkdir(tempDir, os.ModePerm)

	err = unzipFile(zipPath, tempDir)
	if err != nil {
		fmt.Println("Error unzipping file:", err)
		return
	}

	baseProgramPath := filepath.Join(tempDir, "base")

	// Platform-specific executable extension
	executableExt := ""
	if runtime.GOOS == "windows" {
		executableExt = ".exe"
	}

	// Rename the executable
	newExecPath := execPath + "_new" + executableExt
	err = os.Rename(baseProgramPath, newExecPath)
	if err != nil {
		fmt.Println("Error replacing executable:", err)
		return
	}

	// Delete the old executable
	err = os.Remove(execPath)
	if err != nil {
		fmt.Println("Error removing old executable:", err)
		return
	}

	// Rename the new executable to the original name
	err = os.Rename(newExecPath, execPath)
	if err != nil {
		fmt.Println("Error renaming new executable:", err)
		return
	}

	fmt.Println("Update successful. Restarting...")

	restart()
}
