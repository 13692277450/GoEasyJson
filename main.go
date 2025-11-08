package main

/*
Version: 0.01
Author: Mang Zhang, Shenzhen China
Release Date: 2025-11-08
Project Name: GoEasyJson
Description: A tool to help fast and automatically create json server.
Copy Rights: MIT License
Email: m13692277450@outlook.com
Mobile: +86-13692277450
HomePage: www.pavogroup.top , github.com/13692277450
*/
import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
)

var (
	router     *mux.Router
	routes     = make(map[string]bool)
	routesLock sync.RWMutex
	port       string
	watcher    *fsnotify.Watcher
)

// excludedExtensions is a map of file extensions to exclude from scanning and serving.

var excludedExtensions = map[string]bool{
	".exe": true,
	".cmd": true,
	".bat": true,
	".msi": true,
	".rar": true,
	".zip": true,
	".7z":  true,
	".log": true,
	".go":  true,
}

var (
	CurrentVersion        = "0.01"
	NewVersionIsAvailable = ""
	IsUpgrade             = flag.Bool("upgrade", false, "Run with -upgrade to upgrade new version of GoEasyJson")
	UpgradeDetails        = ""
	SignalString          string
)

func init() {
	flag.StringVar(&port, "port", "2006", "Server port (default: 2006)")
}

// Scan files and update routes based on JSON files in the current directory.
func scanDirectory() {
	currentDir, err := os.Getwd()
	if err != nil {
		Lg.Errorf("Error getting current directory: %v", err)
		return
	}
	Lg.Infof("Current directory: %s", currentDir)
	Lg.Info("Scanning directory for JSON files...")

	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		Lg.Errorf("Error reading directory: %v", err)
		return
	}

	newRoutes := make(map[string]bool)

	for _, file := range files {
		if file.IsDir() {
			continue // bypass files in directories
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if excludedExtensions[ext] {
			continue // bypass excluded file extensions
		}

		// For JSON file create route

		if strings.ToLower(ext) == ".json" {
			routePath := "/" + strings.TrimSuffix(file.Name(), ext)
			newRoutes[routePath] = true
		}
	}

	// Update routes

	updateRoutes(newRoutes)
}

// Update routes configuration based on new routes.
func updateRoutes(newRoutes map[string]bool) {
	routesLock.Lock()
	defer routesLock.Unlock()

	// Add new routes
	for route := range newRoutes {
		if !routes[route] {
			log.Printf("Adding new route: %s", route)
			Lg.Infof("Adding new route: %s", route)
			router.HandleFunc(route, createFileHandler(route)).Methods("GET")
			routes[route] = true
		}
	}

	// Remote routes that no longer exist
	for route := range routes {
		if !newRoutes[route] {
			Lg.Infof("Route %s no longer exists", route)

			delete(routes, route)
		}
	}
}

// File process
func createFileHandler(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := strings.TrimPrefix(route, "/") + ".json"
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			Lg.Errorf("Error reading file %s: %v", filename, err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Setup response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
		log.Printf("The Json response route %s was responsed success.", filename)
		Lg.Infof("The Json response route %s was responsed success.", filename)
	}
}

// Initialize file watcher and start monitoring for changes.
func initFileWatcher() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Get current directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	err = watcher.Add(currentDir)
	if err != nil {
		return err
	}
	log.Printf("File watcher initialized, monitoring changes in %s", currentDir)
	Lg.Info("File watcher initialized, monitoring changes in current directory")

	// Start file watcher event loop
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("File event detected: %s  %s", event.Name, event.Op)
				Lg.Infof("File event detected: %s  %s", event.Name, event.Op)
				// Detect JSON file changes or creation/deletion of files
				ext := strings.ToLower(filepath.Ext(event.Name))
				if ext == ".json" || event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					// Delay the scan to avoid excessive scanning during rapid changes
					time.AfterFunc(100*time.Millisecond, func() {
						scanDirectory()
					})
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				Lg.Infof("File watcher error: %v", err)
			}
		}
	}()

	return nil
}

func main() {
	flag.Parse()

	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println("GoEasyJson version 0.01, Author: Mang Zhang, Shenzhen, China")
	fmt.Println("Source code: github.com/13692277450/goeasyjson")
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println("GoEasyJson has file watch function to monitor new json files automatically.")
	fmt.Println("Just put your new Json file in the same directory as this \nprogram, and it will be served automatically.")
	fmt.Printf("Your json route will be localhost:%s/filename-without-extension.\n", port)
	fmt.Println("---------------------------------------------------------------------------")
	LogrusConfigInit()
	go NewVersionCheck()
	fmt.Println("GoEasyJson is checking new version and initiallizing, pls wait 3 seconds...")
	fmt.Println("")
	time.Sleep(time.Second * 3)
	SignalString += SignalString + NewVersionIsAvailable // check for new version
	fmt.Println(SignalString)
	fmt.Println(" ")
	if *IsUpgrade {
		DownloadUpgrade() // download new version
		os.Exit(0)
	}
	if _, err := os.Stat("goeasyjson.exe.old"); os.IsNotExist(err) {
	} else {
		os.Remove("goeasyjson.exe.old")
		fmt.Printf("The old version application was removed success.\n")
	}
	// Initialize router
	router = mux.NewRouter()

	// Initialize file watcher and start monitoring for changes.
	err := initFileWatcher()
	if err != nil {
		Lg.Infof("Failed to initialize file watcher: %v", err)
		Lg.Info("Falling back to periodic scanning only")
	} else {
		defer watcher.Close()
	}

	// Start Scan files
	scanDirectory()

	// Setup periodic scanning every 60 seconds as backup solution
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			scanDirectory()
		}
	}()

	// Add health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"status": "ok", "message": "Server is running"}
		log.Printf("Health check endpoint successed/")
		Lg.Info("Health check endpoint successed.")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Start the server
	log.Printf("Starting server on port %s...", port)
	Lg.Infof("Starting server on port %s...", port)
	Lg.Infof("Access JSON files at http://localhost:%s/[filename-without-extension]", port)
	Lg.Info("Server will automatically update routes when JSON files are added/removed/modified")

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		Lg.Infof("Server error: %v", err)
	}
	LogFile.Close()
}
