//Package utils fuctions mainly utility purposes
package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

// CheckError check whether an error occured
func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// // InitLogger setup logger for both printing file and console
// func InitLogger(path string, methodname bool, isstdout bool) {
// 	// remove old file if exist
// 	os.Remove(path)
// 	// set formatter of logger
// 	log.SetFormatter(&log.TextFormatter{
// 		ForceColors:      false,
// 		DisableColors:    true,
// 		DisableTimestamp: true,
// 	})
// 	// open logging file
// 	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
// 	// check error
// 	CheckError(err)
// 	// defer f.Close()
// 	wrt := io.Writer(f)
//
// 	if isstdout {
// 		// log both file and console
// 		wrt = io.MultiWriter(os.Stdout, f)
// 	}
// 	// set output
// 	log.SetOutput(wrt)
// 	// set whether we log filename as well
// 	log.SetReportCaller(methodname)
// 	// log.SetLevel(log.ErrorLevel)
// 	log.Info("Logger has been initilized.")
// }

// InitLogger setup logger for both printing file and console
func InitLogger(path string, methodname bool, isstdout bool) *log.Logger {
	// remove old file if exist
	os.Remove(path)

	logger := log.New()
	// set formatter of logger
	logger.SetFormatter(&log.TextFormatter{
		ForceColors:      false,
		DisableColors:    true,
		DisableTimestamp: true,
	})
	// open logging file
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	// check error
	CheckError(err)
	// defer f.Close()
	wrt := io.Writer(f)

	if isstdout {
		// log both file and console
		wrt = io.MultiWriter(os.Stdout, f)
	}
	// set output
	logger.SetOutput(wrt)
	// set whether we log filename as well
	logger.SetReportCaller(methodname)
	// log.SetLevel(log.ErrorLevel)
	logger.Info("Logger has been initilized.")

	return logger
}

// ReadConfFile set config file variables and read it
// it uses TOML format for config file
func ReadConfFile() {
	viper.SetConfigName("config") // no need to include file extension
	viper.AddConfigPath("../")    // set the path of your config file
	viper.AddConfigPath(".")      // set the path of your config file

	err := viper.ReadInConfig()
	CheckError(err)
}

// ###########################################3
// identifyPanic get panic method and line number
func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}

// RecoverPanic recover from panic
func RecoverPanic() {
	log.Info("Recovered error")
	log.Error(identifyPanic())

	r := recover()
	if r == nil {
		return
	}
	log.Error(r)
	// panic("need handle")
}

// Abs find absolute value of an int
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
