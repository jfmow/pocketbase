package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable path:", err)
	}

	// Get the directory containing the executable
	execDir := filepath.Dir(executable)

	// Go up one directory
	newBaseDir := filepath.Join(execDir, "base")
	parentDir := filepath.Join(execDir, "..")

	// Join with "base" to get the new path
	runningAppPath := filepath.Join(parentDir, "base")

	fmt.Println("Running App Path:", runningAppPath)

	// Remove the existing executable
	err = os.Remove(runningAppPath)
	if err != nil {
		log.Fatal("Failed to delete current exe: ", err)
	}

	// Copy the new executable
	copyFile(newBaseDir, runningAppPath)

	// Set permissions for the new executable
	err = os.Chmod(runningAppPath, 0755)
	if err != nil {
		log.Fatal("Failed to set permissions for the new exe: ", err)
	}

	// Run the new executable
	cmd := exec.Command(runningAppPath)

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func copyFile(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
