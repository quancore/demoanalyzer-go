//Package utils fuctions mainly utility purposes
package utils

import (
	"fmt"
	"go/build"
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

// GetGoPath get current go path in the system
func GetGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
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
func InitLogger(path, logLevel string, methodname bool, isstdout bool) *log.Logger {
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
	// parse string, this is built-in feature of logrus
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	logger.SetLevel(ll)
	level := logger.GetLevel()
	logger.WithFields(log.Fields{
		"level": level,
	}).Info("Logger has been initilized")

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

// Absf32 find absolute value of an float32
func Absf32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Maxf32 max for float32
// Return false if max is first argument (x)
func Maxf32(x, y float32) (float32, bool) {
	if x < y {
		return y, true
	}
	return x, false
}

// Minf32 min for float32
// Return false if min is first argument (x)
func Minf32(x, y float32) (float32, bool) {
	if x > y {
		return y, true
	}
	return x, false
}

// SafeDivision divide given numbers. If divider 0, return 0
func SafeDivision(number, divider float32) float32 {
	var result float32
	if divider != 0 {
		result = number / divider
	}

	return result
}
