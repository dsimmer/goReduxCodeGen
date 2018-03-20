package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//Requirements

// separate template files for each type of replacements
// Well documented

// for generating flow props - add in basic if not found. Also check for state and add in if not found.
// If default props present and line item not present in props, add to props.
// If action not present in proptypes, add to props.
// if mapstatetoprops not present in proptypes, add to props.

// For generating actions, must look through for connect. If not found, add it in (plus imports and mapstatetoprops and export). If it is there, then check mapdispatchtoprops sub items.

// For generating reducers, look through for same action names plus the suffix Reducer. Add any not found
// Then look through the combined reducer for the reducers. Add any not found

// To mark an action as custom (and not able to be generated), prefix it with //@NoGen
// To mark a file as custom (and not able to be generated), put //@NoGenFile in the file somewhere
// Change the template if you want the custom tags (generated by, etc)

// template is passed variable types for each action and the action names - otherwise same format as gogen/yacc

// example tags
// Generated by GoReactGen

// End generated code

// separate function that generates flow type from json or from a go struct (with json annotations)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseActions(path, config configuration) error {
	return nil
}

func propTypesGenerator(file string, path string, config configuration) error {
	isPropTypes := strings.Contains(file, config.propTypesMarker)
	if isPropTypes {
		//insert
	} else {
		// create from scratch
	}
	return nil
}
func stateTypesGenerator(file string, path string, config configuration) error {
	isStateTypes := strings.Contains(file, config.stateTypesMarker)
	if isStateTypes {
		//insert
	} else {
		// create from scratch
	}
	return nil
}
func mapDispatchGenerator(file string, path string, actionPath string, config configuration) error {
	isMapDispatch := strings.Contains(file, config.mapDispatchMarker)
	if isMapDispatch {
		// insert lines
	} else {
		//create a new one
	}
	return nil
}
func blankMapStateGenerator(file string, path string, actionPath string, config configuration) error {
	return nil
}
func reducerGenerator(path string, reducerPath string, config configuration) error {
	// Check if any existing reducers in the reducers file and then execute logic
	return nil
}
func combinedReducerGenerator(path string, reducerPath string, config configuration) error {
	// Check if combined reducers exists in the reducers file and then execute logic
	return nil
}

type configuration struct {
	reactMarker        string
	flowMarker         string
	stateMarker        string
	stateTypesMarker   string
	propTypesMarker    string
	defaultPropsMarker string
	reduxMarker        string
	mapDispatchMarker  string
	mapStateMarker     string
	actionPrefix       string
	actionSuffix       string
	reducerPrefix      string
	reducerSuffix      string
}

func main() {
	dir, err := os.Getwd()
	check(err)

	log.Println(dir)

	searchDir := dir

	generatePropTypes := true
	generateStateTypes := true
	generateMapDispatch := true
	generateBlankMapState := true

	generateReducers := true
	generateCombinedReducer := true

	config := configuration{
		reactMarker:        "import React from ",
		flowMarker:         "@flow",
		stateMarker:        "state =",
		stateTypesMarker:   "State:",
		propTypesMarker:    "Props:",
		defaultPropsMarker: "defaultProps:",
		reduxMarker:        "import {connect} from ",
		mapDispatchMarker:  "mapDispatchToProps",
		mapStateMarker:     "mapStateToProps",
		actionPrefix:       "",
		actionSuffix:       "Actions",
		reducerPrefix:      "",
		reducerSuffix:      "Reducer",
	}
	// import markers etc

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		check(err)
		if info.IsDir() {
			return nil
		}
		contents, err := ioutil.ReadFile(path)

		check(err)

		processed := string(contents)

		isReact := strings.Contains(processed, config.reactMarker)
		isFlow := strings.Contains(processed, config.flowMarker)
		isRedux := strings.Contains(processed, config.reduxMarker)

		ActionFile := filepath.Base(path) + string(os.PathSeparator) + config.actionPrefix + info.Name() + config.actionSuffix
		_, err = os.Stat(ActionFile)
		isActionExist := err == nil
		ReducerFile := filepath.Base(path) + string(os.PathSeparator) + config.reducerPrefix + info.Name() + config.reducerSuffix
		_, err = os.Stat(ReducerFile)
		isReducerExist := (err == nil)

		if isFlow && isReact {
			isDefaultProps := strings.Contains(processed, config.defaultPropsMarker)
			isState := strings.Contains(processed, config.stateMarker)

			if generatePropTypes && isDefaultProps {
				propTypesGenerator(processed, path, config)
			}
			if generateStateTypes && isState {
				stateTypesGenerator(processed, path, config)
			}
		}
		if isReact && isRedux && isActionExist {
			isMapState := strings.Contains(processed, config.mapStateMarker)
			if generateMapDispatch {
				mapDispatchGenerator(processed, path, ActionFile, config)
			}
			if generateBlankMapState && !isMapState {
				blankMapStateGenerator(processed, path, ActionFile, config)
			}
		}
		if isActionExist && isReducerExist {
			if generateCombinedReducer {
				combinedReducerGenerator(path, ReducerFile, config)
			}
			if generateReducers {
				reducerGenerator(path, ReducerFile, config)
			}
		}

		// append to files changed which should be returned
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}

	return fileList, nil
}
