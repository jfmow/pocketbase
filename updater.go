package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func updater() {
	// Hardcoded URL to the zip file
	url := "https://raw.githubusercontent.com/jfmow/pocketbase/master/base.zip"

	// Download the zip file
	fmt.Println("Downloading file...")
	executable, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable path:", err)
	}

	filePath := filepath.Join(filepath.Dir(executable), "downloaded.zip")
	uzFilePath := filepath.Join(filepath.Dir(executable), "unzipped")
	err = downloadFile(filePath, url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	fmt.Println("File downloaded successfully.")

	// Unzip the file
	fmt.Println("Unzipping file...")
	err = unzip(filePath, uzFilePath)
	if err != nil {
		fmt.Println("Error unzipping file:", err)
		return
	}
	fmt.Println("File unzipped successfully.")
	err = os.Remove(filePath)
	if err != nil {
		fmt.Println("Error deleting zip file:", err)
		return
	}

	// Run the executable
	fmt.Println("Running base executable...")
	exePath := filepath.Join(uzFilePath, "installer")
	err = os.Chmod(exePath, 0755)
	if err != nil {
		fmt.Println("error setting execution permission:", err)
		return
	}
	err = runExecutable(exePath)
	if err != nil {
		fmt.Println("Error running executable:", err)
		return
	}
	fmt.Println("Executable ran successfully.")
	fmt.Println("Stopping the database!")
	os.Exit(0)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src string, dest string) error {
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
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func runExecutable(executablePath string) error {
	cmd := exec.Command(executablePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
}
