package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var Lg = logrus.New()
var LogFile *os.File

func LogrusConfigInit() {

	LogFile, err := os.OpenFile("GoEasyJson.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		Lg.Fatalf("Can not open log file %v !", err)
		log.Printf("Can not open log file %v !", err)

	}
	//mw := io.MultiWriter(os.Stdout, LogFile)
	//Lg.SetOutput(mw)
	Lg.SetOutput(LogFile)
	Lg.SetFormatter(&logrus.JSONFormatter{})
	Lg.SetLevel(logrus.InfoLevel)
	Lg.Info("Logrus initiallized and started....")
	log.Println("Logrus initiallized and started....")

}
