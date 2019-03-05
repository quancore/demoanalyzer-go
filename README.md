# CS:GO Demo analyzer

CS:GO demo analyzer written in golang. It is used a [parser](https://github.com/markus-wa/demoinfocs-golang) written in golang as well. 


## Purpose

The analyzer has mainly written for a master project about prediction of cs:go player rankings. The purpose of the analyzer is extracting useful features that can summerize the game as good as it can. There will be addding more advanced features to indicate team strategies as well as complicated interactions within players. 
 
## Requirements

This library is intended to be used with  `go 1.11`  or higher as it is built using Go modules.

It's recommended to use modules for consumers as well if possible. If you are unfamiliar with Go modules there's a  [list of recommended resources](https://github.com/markus-wa/demoinfocs-golang/wiki/Go-Modules#recommended-links--articles)  in the wiki.

## Installation

Simply do:  `go get -u github.com/quancore/demoanalyzer-go`

## Running of analyzer

Simply indicate path of the demofile, path of output file and path of log file and then run the main package. It will create normalized features for each player in a text file. Example given below:

    go build
    ./demoanalyzer-go --demofilepath natus-vincere-vs-avangar-m2-train.dem --outpath stat.txt --checkanalyzer --logfilepath log.txt

## Bug report

It is highly possible that the analyzer can fail to analyze a demofile. Because demofiles are very messy, there is no guarantee that this analyzer can analyze all kind of demo files. If you encounter an runtime error or get wrong statistics of a player, simply do: 
 - Create an issue indicating what is the problem (runtime error, statistical error etc.).
 - Put the related demo file url and the result page of the match if possible.
 - Put the log file created by the analyzer.

## Known issues

 - The analyzer cannot analyze very old demo files because of limitation of relaying parser.
 - There is a statistical error for finding correct number of flash assist because of relaying parser.
 - KAST and clutch number can be different a bit compared to HLTV statistics because of different definition of the features in analyzer compared to matchmaking server.
