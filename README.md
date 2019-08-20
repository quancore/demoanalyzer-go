# CS:GO Demo analyzer

CS:GO demo analyzer written in golang. It is used a [parser](https://github.com/markus-wa/demoinfocs-golang) written in golang as well. 


## Purpose

The analyzer has mainly written for a master project about prediction of cs:go player rankings. The purpose of the analyzer is extracting useful features that can summerize the game as good as it can. There will be addding more advanced features to indicate team strategies as well as complicated interactions within players.

Link to master thesis: https://drive.google.com/open?id=1JXIB57BA2XBTYVLSy6Xg_5nfL6dWyDmG

## How it works?

You can check how it works by checking **Demo file analysis** section of my master thesis.
 
## Requirements

This library is intended to be used with  `go 1.11`  or higher as it is built using Go modules.

It's recommended to use modules for consumers as well if possible. If you are unfamiliar with Go modules there's a  [list of recommended resources](https://github.com/markus-wa/demoinfocs-golang/wiki/Go-Modules#recommended-links--articles)  in the wiki.

## Installation

Simply do:  `go get -u github.com/quancore/demoanalyzer-go`

## Commands of analyzer

Simply indicate path of the demofile, path of output file and path of log file and then run the main package. It will create normalized features for each player in a text file. Example given below:

    go build
    ./demoanalyzer-go --demofilepath natus-vincere-vs-avangar-m2-train.dem --outpath stat.txt --checkanalyzer --logfilepath log.txt

`--checkanalyzer`: Very useful flag for checking the results of analyzer. It is very helpful to find out whether a demo file has been analyzer correctly.

Example command to build:
`go build -o ./bin/demoanalyzer-go`

Example command to print all events using underlying parser
`go run -tags debugdemoinfocs examples/print-events/print_events.go -demo /paath/to/demofile > out.log`

Example command to run test cases:
`cd test_analyser/`
`go test -v -short -timeout 9000s> out.log`

Example command to run:
`./demoanalyzer-go --demofilepath /path/to/demofile --outpath /path/to/outfile --checkanalyzer --logfilepath /path/to/logfile`

## Feature request

It is highly appreciated to introduce new game features to summerize players and games better. If you think a feature can add value, you can create a PR with indication of importance of feature and exact definition of it.

## How to add a new feature?

### Addition to *player* class:

    

-   Add variables and getter/setter methods for a new feature to *player.go*.
   
-   Add related variables to *ResetPlayerState(player.go)* to reset player feature value for a match start or for the second parsing stage.
   
-   Add related variables to *OutputPlayerState(player.go)*  to output related variable to a text file at the end of the analyzing stage.
   
-   Append your feature name to *features* string in *config.toml*.  *features* string is used for the feature name header in text output. Make sure the order of appending of your feature in *OutputPlayerState* method is the same with the order of your feature name in *features* string.
   
-   You can increase the version of analyzer since you have modified the analyzer using *analyzer_version* variable in *config.toml.*
    

### Addition to event notification:
-   If your event is no need to a **pre or post check** (please read the last paragraph of demo file analysis section in the master thesis), create an event handler method in the related event handler module (if the feature is about players, use *player_event_handlers.go*, else the event is about whole match use *match_event_handlers.go*), and register this handler in *dispatchPlayerEvents*( no need for a match feature). You can find the event list and explanation of each event [here](https://godoc.org/github.com/markus-wa/demoinfocs-golang/events) . If your event handler needs a global instance variable or constant to keep some records, you can add any variable to *Analyser(analyser.go)* struct.
   
-   Register related event to your event handler in *registerPlayerEventHandlers*(or *registerMatchEventHandlers*).
   
-   If your event is needed to **a pre or post check**, in addition to the steps above, you need to schedule pre/post-event checkers using the scheduler. Scheduling these checkers have been done on the first parsing stage. First, create pre/post-event checker struct and related methods in *custom_events.go* (example custom events have been placed). Then, create a method in *player_check_event_handlers.go* to register this checker event to the scheduler (or you can add these checker events in any place because the scheduler reference has been set up in *Analyzer* struct. But be careful that you need to ensure that your custom checker has been registered to the scheduler in the first parsing stage. On the second parsing stage, all checker events have been scheduled and resolved). Both one time checkers (for example checker for a kill event) and periodic checkers (for example map place control checker has been executed periodically) have been supported.
    
### Test your feature:

    

-   You can add print statements using the logger to ensure your handler has been triggered correctly and the context of the handler is expected.
   
-   You can compare analyzer result with real results from a matchmaking server to check your feature outputs correctly.
   
-   Be careful about nil pointer checking on event handlers because it is highly probable that parser event can be emitted with fields includes nil values/pointers. I have created several event checkers for different situations in *checkers.go*.
   

- After ensuring your new feature is semantically and syntactically correct, you can run test cases on working demo files to check whether analyzer analyzes all already working demo files without any runtime error (to test: navigate to *test_analyzer* directory and run **go test -v -short -timeout 9000s> out.log**).  All test cases have been placed in *test_cases.go*, and you can add more test cases. It will analyze all working demo files in the directory that you set (*demofile_path* in *config.toml*).

## Bug report

It is highly possible that the analyzer can fail to analyze a demofile. Because demofiles are very messy, there is no guarantee that this analyzer can analyze all kind of demo files. If you encounter an runtime error or get wrong statistics of a player, simply do: 
 - Create an issue indicating what is the problem (runtime error, statistical error etc.).
 - Put the related demo file url and the result page of the match if possible.
 - Change log level to debug using TOML file and put the log file created by the analyzer.

## Known issues

 - The analyzer cannot analyze very old demo files because of limitation of relaying parser.
 - There is a statistical error for finding correct number of flash assist because of relaying parser.
 - KAST and clutch number can be different a bit compared to HLTV statistics because of different definition of the features in analyzer compared to matchmaking server.
