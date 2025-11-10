package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
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

	filepathWindows := "./goeasyjson.exe"
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

	fmt.Println("Update completed successfully. Old version saved as " + oldFile + ", the old version will be removed automatically when application launch next time.")
	os.Exit(0) // Exit the program after successful update
}

func DownlaodOption() {
	sysType := runtime.GOOS
	if sysType == "windows" {
		DownloadUrl = "http://www.pavogroup.top/software/goeasyjson/goeasyjson.exe"
		DownloadUpgradeWindows(DownloadUrl)
		//DownloadWithBar(DownloadUrl)
	} else {
		DownloadUrl = "http://www.pavogroup.top/software/goeasyjson/goeasyjsonLinuxVersion"
		DownloadUpgradeLinux(DownloadUrl)
	}
}

func DownloadUpgradeLinux(DownloadUrl string) {
	filepathLinux := "./goeasyjsonLinuxVersion"
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

	fmt.Println("Update completed successfully. Old version saved as " + oldFile + ", the old version will be removed automatically when application launch next time.")
	os.Exit(0) // Exit the program after successful update
}

func DownloadWithBar(url string) {

	//url = "http://www.pavogroup.top/software/goeasyjson/goeasyjson.exe"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	if resp == nil || resp.Body == nil {
		fmt.Println("Empty response received")
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile("goeasyjson1.exe", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer f.Close()
	//green := color.New(color.FgGreen).SprintFunc()
	//yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	//purple := color.New(color.FgMagenta).SprintFunc()
	progressbar.OptionSetTheme(progressbar.Theme{
		Saucer:        cyan("#"),
		SaucerHead:    cyan("#"),
		SaucerPadding: " ",
		BarStart:      "[",
		BarEnd:        "]",
	})
	progressbar.OptionSetDescription("[cyan][1/3][reset] Writing moshable file...")

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		Cyan.Render("Downloading..."),
	)

	if _, err := io.Copy(io.MultiWriter(f, bar), resp.Body); err != nil {
		fmt.Printf("\nDownload failed: %v\n", err)
		return
	}
}
