package testAnalyser

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"testing"

	analyser "github.com/quancore/demoanalyzer-go/analyser"
	utils "github.com/quancore/demoanalyzer-go/utils"
	logging "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var numConcurrentWorker int
var demofilePath string
var logPrefix string
var logLevel string
var outputPrefix string
var isMethodName bool
var stdout bool
var log *logging.Logger

func init() {
	utils.ReadConfFile()

	// get conf variables
	demofilePath = viper.GetString("test.demofile_path")
	numConcurrentWorker = viper.GetInt("test.concurrent_worker")
	logPrefix = viper.GetString("test.log_prefix")
	logLevel = viper.GetString("test.log_level")
	outputPrefix = viper.GetString("test.output_prefix")
	isMethodName = viper.GetBool("log.is_method_name")
	stdout = viper.GetBool("log.stdout")

	viper.Set("checkanalyzer", true)

	// init test logger
	log = utils.InitLogger("test_log.txt", logLevel, isMethodName, true)

	log.Info("Logger has been initilized")

}

//TestDemofiles test correctness of analyser with demofiles
func TestDemofiles(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	log.Info(fmt.Sprintf("gopath: %s", gopath))

	demofilePath = path.Join(gopath, demofilePath)

	// first read all demofiles from dir
	files, err := ioutil.ReadDir(demofilePath)
	if err != nil {
		t.Fatal(err)
	}

	log.Info(fmt.Sprintf("Found files number: %d", len(files)))

	var tasks []*Task

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".dem" {
			filename := file.Name()
			filepath := filepath.Join(demofilePath, filename)
			tasks = append(tasks, NewTask(func() error {
				err := analyseDemofile(filename, filepath, t)
				return err
			}))
		}
	}

	p := NewPool(tasks, numConcurrentWorker)
	p.Run()

}

func analyseDemofile(filename, filepath string, t *testing.T) error {
	// format filename
	filename = strings.Split(filename, ".")[0]
	log.Info(fmt.Sprintf("Now parsing: %s (%s)", filename, filepath))

	// init path variables
	logFilePath := logPrefix + "_" + filename + ".txt"
	outputPath := outputPrefix + "_" + filename + ".txt"

	log.Info(fmt.Sprintf("Log filename: %s out filename: %s", logFilePath, outputPath))

	// init loggers
	// utils.InitLogger(logFilePath, isMethodName, false)

	f, err := os.Open(filepath)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	// initilize analyser
	analyser := analyser.NewAnalyser(f, logFilePath, outputPath, isMethodName, stdout)
	// finally parse demofile
	analyser.Analyze()

	return err

}
