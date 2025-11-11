package main

/*
Version: 0.03
Author: Mang Zhang, Shenzhen China
Release Date: 2025-11-08
Project Name: GoEasyJson
Description: A tool to help fast and automatically create json server and generate test data.
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
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
)

var (
	router     *mux.Router
	routes     = make(map[string]bool)
	routesLock sync.RWMutex
	port       int
	watcher    *fsnotify.Watcher
	genjson    string
	out        string
	qty        int
)

var Red = lipgloss.NewStyle().Foreground(lipgloss.Color("#b507eaff"))
var LightGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("#61f882ff"))
var LightYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#f6f349ff"))

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
	CurrentVersion = "0.0.3"
	IsUpgrade      = flag.Bool("upgrade", false, "Run with -upgrade to update to new version of GoEasyJson")
	// GenJson        = flag.String("genjson", "", "Generate test JSON data from sample file (e.g. -genjson sample.json -out test.json -qty 1000)")
	// OutputFile     = flag.String("out", "", "Output file for generated JSON data")
	// Quantity       = flag.Int("qty", 0, "Number of records to generate")
)

func init() {
	flag.StringVar(&genjson, "genjson", "", "Generate test JSON data from sample file (e.g. -genjson sample.json -out test.json -qty 1000)")
	flag.StringVar(&out, "out", "", "Output file for generated JSON data")
	flag.IntVar(&qty, "qty", 0, "Number of records to generate")
	flag.IntVar(&port, "port", 2006, "Server port (e.g. goeasyjson -port 2006)")

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

// Initialize file watcher and start monitoring for JSON files only.
func initFileWatcher() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}

	// Get current directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Add watch for current directory
	err = watcher.Add(currentDir)
	if err != nil {
		return fmt.Errorf("failed to add watch for directory %s: %v", currentDir, err)
	}

	log.Printf("File watcher initialized, monitoring directory %s for JSON files only", currentDir)
	Lg.Info("File watcher initialized, monitoring directory for JSON files only")
	lastEvent := make(map[string]fsnotify.Op)

	// Start file watcher event loop
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Get file names and converto lowercase extension
				ext := strings.ToLower(filepath.Ext(event.Name))
				// Only porcess JSON files
				if ext == ".json" {
					switch event.Op {
					case fsnotify.Create:
						// Make sure it's a file and not a directory
						if info, err := os.Stat(event.Name); err == nil {
							if !info.IsDir() {
								// New json file was created, add watch for it
								err := watcher.Add(event.Name)
								if err != nil {
									Lg.Errorf("Failed to add watch for new JSON file %s: %v", event.Name, err)
									log.Printf("Failed to add watch for new JSON file %s: %v", event.Name, err)

								} else {
									Lg.Infof("Added watch for new JSON file: %s", event.Name)
									log.Printf("Added watch for new JSON file: %s", event.Name)
									scanDirectory()
								}
							}
						}

					case fsnotify.Write:
						if lastOp, exists := lastEvent[event.Name]; !exists || lastOp != event.Op {
							lastEvent[event.Name] = event.Op
						} else {
							Lg.Infof("Detect changes in JSON file: %s", event.Name)
							log.Printf("Detected changes in JSON file: %s", event.Name)
						}
						// delayed scan to avoid rapid consecutive changes
						time.AfterFunc(600*time.Millisecond, func() {
							scanDirectory()
						})
					case fsnotify.Remove:
						if lastOp, exists := lastEvent[event.Name]; !exists || lastOp != event.Op {
							lastEvent[event.Name] = event.Op
						} else {
							Lg.Infof("JSON file removed: %s", event.Name)
							log.Printf("JSON file removed: %s", event.Name)
						}
						// json file was removed, remove watch
						scanDirectory()
					case fsnotify.Rename:
						if lastOp, exists := lastEvent[event.Name]; !exists || lastOp != event.Op {
							lastEvent[event.Name] = event.Op
						} else {
							Lg.Infof("JSON file renamed: %s", event.Name)
							log.Printf("JSON file renamed: %s", event.Name)
						}
						scanDirectory()
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					// Process new folder
					if info, err := os.Stat(event.Name); err == nil {
						if info.IsDir() {
							// Create watch for new directory
							err := watcher.Add(event.Name)
							if err != nil {
								Lg.Warnf("Failed to add watch for new directory %s: %v", event.Name, err)
								log.Printf("Failed to add watch for new directory %s: %v", event.Name, err)
							}
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				Lg.Errorf("File watcher error: %v", err)
			}
		}
	}()

	return nil
}
func main() {
	flag.Parse()

	// Check if we need to generate JSON data
	if genjson != "" && out != "" && qty > 0 {
		generateTestData(genjson, out, qty)
		return
	}

	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println(LightGreen.Render("GoEasyJson version 0.0.3(11/11/2025), Author: Mang Zhang, Shenzhen, China"))
	fmt.Println(LightGreen.Render("Source code: github.com/13692277450/goeasyjson, HomePage: www.pavogroup.top"))
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println("GoEasyJson has file watch function to monitor new json files automatically.")
	fmt.Println("Just put your new Json file in the same directory as this program,\nit will be served automatically.")
	fmt.Println("")
	fmt.Println(Red.Render("Fake data generator: goeasyjson -genjson sample.json -out test.json -qty 1000."))
	fmt.Println(Red.Render("Customize API port: goeasyjson -port 2006."))
	fmt.Println(Red.Render("Upgrade to new version: goeasyjson -upgrade."))

	fmt.Println("---------------------------------------------------------------------------")

	LogrusConfigInit()
	go NewVersionCheck()
	fmt.Println(Cyan.Render("GoEasyJson is checking new version and initiallizing, pls wait 3 seconds..."))
	time.Sleep(time.Second * 3)
	fmt.Println(LightYellow.Render(NewVersionIsAvailable))
	fmt.Println("")
	if *IsUpgrade {
		DownlaodOption() // download new version
		os.Exit(0)
	}
	if _, err := os.Stat("goeasyjson.exe.old"); os.IsNotExist(err) {
	} else {
		os.Remove("goeasyjson.exe.old")
		fmt.Printf("The old version application was removed success.\n")
	}
	if _, err := os.Stat("goeasyjsonLinuxVersion.old"); os.IsNotExist(err) {
	} else {
		os.Remove("goeasyjsonLinuxVersion.old")
		fmt.Printf("The old version application was removed success.\n")
	}
	if _, err := os.Stat("goeasyjsonLinuxVersion.old"); os.IsNotExist(err) {
	} else {
		os.Remove("goeasyjsonMacVersion.old")
		fmt.Printf("The old version application was removed success.\n")
	}
	// Initialize router
	router = mux.NewRouter()

	// Initialize file watcher and start monitoring for changes.
	err := initFileWatcher()
	if err != nil {
		log.Printf("Failed to initialize file watcher: %v", err)
		log.Println("Falling back to periodic scanning only")
	} else {
		defer watcher.Close()
	}

	// Start Scan files
	scanDirectory()

	// Version check channel
	// versionChan := make(chan string)
	// go func() {
	// 	versionChan <- NewVersionCheck()
	// }()

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
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Start the server
	log.Printf("Starting server on port %d...", port)
	var strPort = fmt.Sprintf("Access JSON API at http://localhost:%d/filename-without-extension\n", port)
	fmt.Println(LightYellow.Render(strPort))
	fmt.Println("Server will automatically update routes when JSON files are added/removed/modified")

	if _, err := os.Stat("goeasyjson.exe.old"); os.IsNotExist(err) {
	} else {
		os.Remove("goeasyjson.exe.old")
		fmt.Printf("The old version application was removed success.\n")
	}

	err = http.ListenAndServe(":"+strconv.Itoa(port), router)
	if err != nil {
		log.Printf("Server error: %v", err)
	}
}

func generateTestData(sampleFile, outputFile string, quantity int) {
	// Read sample JSON file
	sampleData, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		log.Fatalf("Error reading sample JSON file: %v", err)
	}

	// Parse sample JSON to get the structure
	var sample interface{}
	if err := json.Unmarshal(sampleData, &sample); err != nil {
		log.Fatalf("Error parsing sample JSON: %v", err)
	}

	// Generate test data
	var results []interface{}
	for i := 0; i < quantity; i++ {
		// Deep copy sample structure for each record
		sampleCopy, _ := json.Marshal(sample)
		var record interface{}
		_ = json.Unmarshal(sampleCopy, &record)
		record = fillDynamic(record, "")
		results = append(results, record)
	}

	// Write generated data to output file
	outputData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling generated data: %v", err)
	}

	if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	fmt.Printf("Successfully generated %d records and saved to %s\n", quantity, outputFile)
}

// Recursively fill any map[string]interface{} or []interface{} with dynamic data
func fillDynamic(v interface{}, key string) interface{} {
	switch vv := v.(type) {
	case map[string]interface{}:
		for k, val := range vv {
			vv[k] = fillDynamic(val, k)
		}
		return vv
	case []interface{}:
		for i, item := range vv {
			vv[i] = fillDynamic(item, "")
		}
		return vv
	case string:
		switch strings.ToLower(key) {
		case "name", "username", "firstname", "lastname":
			return gofakeit.Name()
		case "email":
			return gofakeit.Email()
		case "grade":
			return rand.Intn(100) // Grade between 1 and 100
		case "age":
			return rand.Intn(100) // Age between 1 and 99
		case "gender":
			return gofakeit.Gender()
		case "address":
			return gofakeit.Address()
		case "phone", "telephone":
			return gofakeit.Phone()
		case "city":
			return gofakeit.City()
		case "country":
			return gofakeit.Country()
		case "state":
			return gofakeit.State()
		case "streetaddress":
			return gofakeit.Street()
		case "zipcode", "postcode":
			return gofakeit.Zip()
		case "company":
			return gofakeit.Company()
		case "jobtitle", "title":
			return gofakeit.JobTitle()
		case "date", "dob", "birthdate":
			return gofakeit.Date().Format("2006-01-02")
		case "datetime", "timestamp":
			return gofakeit.Date().Format(time.RFC3339)
		case "url", "website":
			return gofakeit.URL()
		case "color":
			return gofakeit.Color()
		case "uuid", "id":
			return gofakeit.UUID()
		case "latitude":
			return gofakeit.Latitude()
		case "longitude":
			return gofakeit.Longitude()
		case "word":
			return gofakeit.Word()
		case "sentence":
			return gofakeit.Sentence()
		case "paragraph":
			return gofakeit.Paragraph()
		case "creditcardnumber":
			// gofakeit.CreditCardNumber may not be available in all versions; use CreditCard() and return Number field
			cc := gofakeit.CreditCard()
			return cc.Number
		case "creditcardtype":
			return gofakeit.CreditCardType()
		case "creditcardexpirationdate":
			return gofakeit.CreditCardExp()
		case "creditcardholdername":
			// return a realistic person name as card holder
			return gofakeit.Name()

		case "creditcardexpirationmonth":
			return gofakeit.Month()
		case "creditcardexpirationyear":
			return gofakeit.Year()
		case "quantity":
			return rand.Intn(1000) // Quantity between 0 and 999
		case "price", "amount":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Price between 1.00 and 1000.00
		case "currency":
			return gofakeit.CurrencyShort()
		case "total":
			return fmt.Sprintf("%.2f", gofakeit.Price(100, 10000)) // Total between 100.00 and 10000.00
		case "vat":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 100)) // VAT between 1.00 and 100.00
		case "vatrate":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 100)) // VAT rate between 1.00 and 100.00%
		case "discount":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 100)) // Discount between 1.00 and 100.00%
		case "discountamount":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 100)) // Discount amount between 1.00 and 100.00
		case "totalwithvat":
			return fmt.Sprintf("%.2f", gofakeit.Price(100, 10000)) // Total with VAT between 100.00 and 10000.00
		case "totalwithoutvat":
			return fmt.Sprintf("%.2f", gofakeit.Price(100, 10000)) // Total without VAT between 100.00 and 10000.00
		case "vatamount":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 100)) // VAT amount between 1.00 and 100.00
		case "totalwithdiscount":
			return fmt.Sprintf("%.2f", gofakeit.Price(100, 10000)) // Total with Discount between 100.00 and 10000.00
		case "totalwithoutdiscount":
			return fmt.Sprintf("%.2f", gofakeit.Price(100, 10000)) // Total without Discount between 100.00 and 10000.00
		case "kilogram":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Kilogram between 1.00 and 1000.00
		case "gram":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Gram between 1.00 and 1000.00
		case "milligram":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Milligram between 1.00 and 1000.00
		case "liter":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Liter between 1.00 and 1000.00
		case "milliliter":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Milliliter between 1.00 and 1000.00
		case "centimeter":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Centimeter between 1.00 and 1000.00
		case "millimeter":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Millimeter between 1.00 and 1000.00
		case "weight":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Weight between 1.00 and 1000.00
		case "height":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Height between 1.00 and 1000.00
		case "width":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Width between 1.00 and 1000.00
		case "depth":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Depth between 1.00 and 1000.00
		case "ip":
			return gofakeit.IPv4Address()
		case "ipv6":
			return gofakeit.IPv6Address()
		case "macaddress":
			return gofakeit.MacAddress()
		case "mac":
			return gofakeit.MacAddress()
		case "miles":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Miles between 1.00 and 1000.00
		case "kilometers":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Kilometers between 1.00 and 1000.00
		case "feet":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Feet between 1.00 and 1000.00
		case "inches":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Inches between 1.00 and 1000.00
		case "pounds":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Pounds between 1.00 and 1000.00
		case "grams":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Grams between 1.00 and 1000.00
		case "milliliters":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Milliliters between 1.00 and 1000.00
		case "centimeters":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Centimeters between 1.00 and 1000.00
		case "millimeters":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Millimeters between 1.00 and 1000.00
		case "seconds":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Seconds between 1.00 and 1000.00
		case "minutes":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Minutes between 1.00 and 1000.00
		case "hours":

			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Hours between 1.00 and 1000.00
		case "days":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Days between 1.00 and 1000.00
		case "months":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Months between 1.00 and 1000.00
		case "years":
			return fmt.Sprintf("%.2f", gofakeit.Price(1, 1000)) // Years between 1.00 and 1000.00

		case "password":
			return gofakeit.Password(true, true, true, true, true, 12)
			// Add more cases for other fields as needed...
		case "day":
			return gofakeit.Day()
		case "month":
			return gofakeit.Month()
		case "year":
			return gofakeit.Year()
		case "companysuffix":
			return gofakeit.CompanySuffix()
		case "unit":
			return gofakeit.Unit()
		case "area":
			return rand.Float64() * 10000 // Area between 0.00 and 10000.00
		case "street":
			return gofakeit.Street()
		case "rich":
			return gofakeit.Bool()

		default:
			return gofakeit.Word()
		}
	// case float64:
	// 	return rand.Float64()
	case bool:
		return rand.Intn(2) == 1
	default:
		return generateRandomValue(key)
	}
}

func generateRandomValue(key string) interface{} {
	// Generate a random number and append to key name
	randNum := rand.Intn(9000000) + 1000000 // 7-digit random number
	return fmt.Sprintf("%s%d", key, randNum)
}
