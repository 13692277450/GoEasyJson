package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cheggaaa/pb/v3"
)

var Cyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
var DownloadUrl string

func downloadFile(url, filepath string) error {

	// HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Download new file was failure.")

		return err
	}
	defer resp.Body.Close()

	// Get file size
	size := resp.ContentLength

	// creat progress bar
	bar := pb.Full.Start64(size)
	defer bar.Finish()

	// create temp file first
	//tempFile := filepathWindows + ".tmp"
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println("Create tmp file was failure.")

		return err
	}
	defer file.Close()

	// create writer with progress bar
	writer := bar.NewProxyWriter(file)

	// write file
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		fmt.Println("Create new file was failure.")
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

// DownloadUpgrade
func DownloadUpgradeWindows(DownloadUrl string) {

	//filepathWindows := "./goeasyjson.exe"
	filepathWindows := filepath.Base(os.Args[0])

	go func() {
		fmt.Println(Cyan.Render("Starting download upgrade from: ", DownloadUrl+"\n"))
		for i := 1; i < 15; i++ {
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
		}
	}()
	tempFile := filepathWindows + ".tmp" // save version as .tmp file
	err := downloadFile(DownloadUrl, tempFile)
	if err != nil {
		fmt.Println("Download new file error.")
		return
	}

	// Rename current executable to .old and rename the new one to current executable
	oldFile := filepathWindows + ".old"
	os.Remove(oldFile) // Remove old backup if exists
	if err := os.Rename(filepathWindows, oldFile); err != nil {
		fmt.Println("Move to old file was failure.")
		return
	}

	// Create update batch script
	batchContent := `@echo off
timeout /t 2 /nobreak >nul
move /Y "` + tempFile + `" "` + filepathWindows + `"
del "%~f0"
`
	batchFile := "update.bat"
	if err := os.WriteFile(batchFile, []byte(batchContent), 0755); err != nil {
		fmt.Println("Running batch file was failure.")
		return
	}
	// Run the batch file and wait for completion
	fmt.Println("Executing update script...")
	cmd := exec.Command("cmd.exe", "/C", batchFile)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing update script: %v\n", err)
		return
	}
	time.Sleep(3 * time.Second)

	// Verify the new file exists
	if _, err := os.Stat(filepathWindows); os.IsNotExist(err) {
		fmt.Printf("Error: New file %s was not created\n", filepathWindows)
		return
	}

	fmt.Println("Update completed successfully. Old version saved as " + oldFile + " which will be removed automatically when application launch next time.")
	os.Exit(0) // Exit the program after successful update
}

func DownlaodOption() {
	sysType := runtime.GOOS
	switch sysType {
	case "windows":
		DownloadUrl = "http://www.pavogroup.top/software/goeasyjson/goeasyjson.exe"
		DownloadUpgradeWindows(DownloadUrl)
		//DownloadWithBar(DownloadUrl)
	case "linux":
		DownloadUrl = "http://www.pavogroup.top/software/goeasyjson/goeasyjsonLinuxVersion"
		DownloadUpgradeLinux(DownloadUrl)
	case "darwin":
		DownloadUrl = "http://www.pavogroup.top/software/goeasyjson/goeasyjsonMacVersion"
		DownloadUpgradeMac(DownloadUrl)
	default:
		fmt.Println("Unsupported OS for upgrade.")
	}
}

func DownloadUpgradeMac(DownloadUrl string) {
	//filepathMac := "./goeasyjsonMacVersion"
	filepathMac := filepath.Base(os.Args[0])
	go func() {
		fmt.Println(Cyan.Render("Starting download upgrade from: ", DownloadUrl+"\n"))
		for i := 1; i < 15; i++ {
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
		}
	}()

	tempFile := filepathMac + ".tmp"
	err := downloadFile(DownloadUrl, tempFile)
	if err != nil {
		fmt.Println("Download new file error.")
		return
	}

	// Rename current executable to .old and rename the new one to current executable
	oldFile := filepathMac + ".old"
	os.Remove(oldFile) // Remove old backup if exists
	if err := os.Rename(filepathMac, oldFile); err != nil {
		// It's okay if the current file doesn't exist (first install)
		fmt.Printf("Move to old file was failure: %v\n", err)
	}

	// Create update shell script for macOS (same behavior as Linux)
	shContent := "#!/bin/sh\n" +
		"sleep 2\n" +
		"mv \"" + tempFile + "\" \"" + filepathMac + "\"\n" +
		"chmod +x \"" + filepathMac + "\"\n" +
		"rm -- \"$0\"\n"

	shFile := "update_mac.sh"
	if err := os.WriteFile(shFile, []byte(shContent), 0755); err != nil {
		fmt.Println("Writing update script was failure.")
		return
	}

	// Run the shell script and wait a moment for it to take over
	fmt.Println("Executing update script...")
	cmd := exec.Command("sh", shFile)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing update script: %v\n", err)
		return
	}
	time.Sleep(3 * time.Second)

	// Verify the new file exists
	if _, err := os.Stat(filepathMac); os.IsNotExist(err) {
		fmt.Printf("Error: New file %s was not created\n", filepathMac)
		return
	}

	fmt.Println("Update completed successfully. Old version saved as " + oldFile + " which will be removed automatically when application launch next time.")
	os.Exit(0) // Exit the program after successful update
}

func DownloadUpgradeLinux(DownloadUrl string) {
	//filepathLinux := "./goeasyjsonLinuxVersion"
	filepathLinux := filepath.Base(os.Args[0])

	go func() {
		fmt.Println(Cyan.Render("Starting download upgrade from: ", DownloadUrl+"\n"))
		for i := 1; i < 15; i++ {
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
		}
	}()

	tempFile := filepathLinux + ".tmp"
	err := downloadFile(DownloadUrl, tempFile)
	if err != nil {
		fmt.Println("Download new file error.")
		return
	}

	// Rename current executable to .old and rename the new one to current executable
	oldFile := filepathLinux + ".old"
	os.Remove(oldFile) // Remove old backup if exists
	if err := os.Rename(filepathLinux, oldFile); err != nil {
		// It's okay if the current file doesn't exist (first install)
		fmt.Printf("Move to old file was failure: %v\n", err)
	}

	// Create update shell script
	shContent := "#!/bin/sh\n" +
		"sleep 2\n" +
		"mv \"" + tempFile + "\" \"" + filepathLinux + "\"\n" +
		"chmod +x \"" + filepathLinux + "\"\n" +
		"rm -- \"$0\"\n"

	shFile := "update.sh"
	if err := os.WriteFile(shFile, []byte(shContent), 0755); err != nil {
		fmt.Println("Writing update script was failure.")
		return
	}

	// Run the shell script and wait a moment for it to take over
	fmt.Println("Executing update script...")
	cmd := exec.Command("sh", shFile)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing update script: %v\n", err)
		return
	}
	time.Sleep(3 * time.Second)

	// Verify the new file exists
	if _, err := os.Stat(filepathLinux); os.IsNotExist(err) {
		fmt.Printf("Error: New file %s was not created\n", filepathLinux)
		return
	}

	fmt.Println("Update completed successfully. Old version saved as " + oldFile + " which will be removed automatically when application launch next time.")
	os.Exit(0) // Exit the program after successful update
}
