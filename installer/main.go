package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	zipURL         = "https://proti.suddsy.dev/api/files/x2v7p9msc2kdas2/od4hxwrl3b7kcha/base_a08AmRBEXo.zip?token=" // Replace with your actual URL
	zipName        = "base.zip"
	targetDir      = "base"
	executableName = "base"
)

func main() {
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

	// Replace the existing executable with the new one
	err = replaceExecutable(filepath.Join(targetDir, executableName))
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
