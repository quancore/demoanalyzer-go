package main

import (
	"fmt"
	"os"

	analyser "github.com/quancore/demoanalyzer-go/analyser"
	utils "github.com/quancore/demoanalyzer-go/common"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	utils.ReadConfFile()

	pflag.Int("concurrentdemos", 1, "The `number` of current demos will be parse")
	pflag.String("demofilepath", "", "The path of demofile")
	pflag.String("outpath", "", "The path of result text file")
	pflag.Bool("checkanalyzer", false, "Flag whether test analyser result when finished")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if _, err := os.Stat(viper.GetString("demofilepath")); err != nil {
		panic(fmt.Sprintf("Failed to read test demo %q", viper.GetString("demofilepath")))
	}
}

func main() {
	// defer utils.RecoverPanic()

	demoFilePath := viper.GetString("demofilepath")
	outPath := viper.GetString("outpath")
	logpath := viper.GetString("log.path")
	isMethodName := viper.GetBool("log.is_method_name")

	f, err := os.Open(demoFilePath)
	defer f.Close()
	utils.CheckError(err)

	// initilize analyser
	analyser := analyser.NewAnalyser(f, logpath, outPath, isMethodName, true)
	// finally parse demofile
	analyser.Analyze()

}
