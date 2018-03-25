package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Weekend: finish and test the components built so far: prop type replacement and dispatch/state map inserts

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
// definitely dont change functions if they use an import
func check(e error) {
	if e != nil {
		panic(e)
	}
}

var typeDict = map[string]string{
	"undefined": `\s*(undefined)\s*`,
	"null":      `\s*(null)\s*`,
	"float":     `^\s*[+-]?[0-9]*[\.][0-9]+$\s*`,
	"int":       `^\s*[+-]?[0-9]+$\s*`,
	"array":     `\s*\[\s*`,
	"string":    `^.*[\D].*$`,
}
var javascriptDict = map[string]string{
	"afterImport": `\s*(undefined)\s*`,
	"null":        `\s*(null)\s*`,
	"float":       `^\s*[+-]?[0-9]*[\.][0-9]+$\s*`,
	"int":         `^\s*[+-]?[0-9]+$\s*`,
	"array":       `\s*\[\s*`,
	"string":      `^.*[\D].*$`,
}

// todo change all splits to regex - so we can get the closing brackets correctly

// TODO: relies on types being before actioncreators - good practice anyways
func parseObject(path string, firstMarker string, lastMarker string, spacer string) ([]string, error) {
	var lines []string

	contents, err := ioutil.ReadFile(path)
	check(err)
	processed := string(contents)
	if !strings.Contains(processed, firstMarker) {
		return lines, errors.New("No firstMarker found")
	}
	processedSplit := strings.SplitAfter(processed, firstMarker)

	processedFirst := strings.SplitAfter(processedSplit[1], lastMarker)
	processedLast := strings.SplitAfter(processedFirst[0], "/n")

	for _, line := range processedLast {
		if strings.Contains(line, lastMarker) {
			break
		} else {
			lines = append(lines, strings.SplitAfter(strings.SplitAfter(line, spacer)[0], ":")[0])
		}
	}

	return lines, nil
}

// maybe just operate on an index?

func parseObjectValues(path string, firstMarker string, lastMarker string, spacer string, replacer func(string) string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	check(err)
	processed := string(contents)
	if !strings.Contains(processed, firstMarker) {
		return "", errors.New("No firstMarker found")
	}
	processedSplit := strings.SplitAfter(processed, firstMarker)

	processedFirst := strings.SplitAfter(processedSplit[1], lastMarker)
	processedLast := strings.SplitAfter(processedFirst[0], "/n")

	for index, line := range processedLast {
		if strings.Contains(line, lastMarker) {
			break
		} else {
			value := strings.SplitAfter(strings.SplitAfter(line, spacer)[0], ":")
			value[1] = replacer(value[1])
			processedLast[index] = strings.Join(value, "/n")
		}
	}

	// join string up again
	file := strings.Join(processedLast, "/n")

	return file, nil
}

// TODO: relies on types being before actioncreators - good practice anyways
func parseActions(path string, config configuration) ([]string, []string, []string, error) {
	var actions []string
	var types []string
	var APITypes []string

	contents, err := ioutil.ReadFile(path)
	check(err)
	processed := string(contents)
	if !strings.Contains(processed, config.actionCreatorMarker) {
		return types, APITypes, actions, errors.New("No actionCreator found")
	}
	processedSplit := strings.SplitAfter(processed, config.actionCreatorMarker)
	// identifier is every type line until }
	processedTypes := strings.SplitAfter(processedSplit[0], "types = {")
	processedTypes = strings.SplitAfter(processedTypes[1], "/n")

	for _, line := range processedTypes {
		if strings.Contains(line, "}") {
			break
		} else if strings.Contains(line, "_REQUEST") {
			APITypes = append(APITypes, strings.SplitAfter(strings.SplitAfter(line, "'")[1], "_REQUEST")[0])
		} else if strings.Contains(line, "_REPLY") || strings.Contains(line, "_ERROR") {
			continue
		} else {
			types = append(types, strings.SplitAfter(line, "'")[1])
		}
	}

	// identifier is any line prefixed by 4 spaces (setting and no more)
	// then get name

	processedAC := strings.SplitAfter(processedSplit[1], "/n")
	for _, line := range processedAC {
		processLine, err := regexp.Match("", []byte(line))
		check(err)
		if processLine && !strings.Contains(line, "}") {
			actions = append(actions, strings.SplitAfter(strings.SplitAfter(line, config.spacer)[1], ":")[0])
		}
	}

	return types, APITypes, actions, nil
}

func replaceTypes(string) string {
	return ""
}

func propTypesGenerator(file string, path string, config configuration) error {
	isPropTypes := strings.Contains(file, config.propTypesMarker)
	if isPropTypes {
		// Not needed?
		// existingProps, err := parseObject(path, config.propTypesMarker, "}", config.spacer)
		// check(err)
		//insert
	} else {
		// create from scratch
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindIndex([]byte(file))[0]
		snippet, err := parseObjectValues(path, config.defaultPropsMarker, "}", config.spacer, replaceTypes)
		snippet = "type PropTypes: {/n" + snippet + "/n};"
		check(err)
		newFile := file[:loc] + snippet + file[loc:]
		err = ioutil.WriteFile(path, []byte(newFile), 0644)
		check(err)
	}
	return nil
}
func stateTypesGenerator(file string, path string, config configuration) error {
	isStateTypes := strings.Contains(file, config.stateTypesMarker)
	if isStateTypes {
		// not needed?
		// existingStates, err := parseObject(path, config.stateTypesMarker, "}", config.spacer)
		// check(err)
		//insert
	} else {
		// create from scratch
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindIndex([]byte(file))[0]
		snippet, err := parseObjectValues(path, config.stateMarker, "}", config.spacer, replaceTypes)
		snippet = "type StateTypes: {/n" + snippet + "/n};"
		check(err)
		newFile := file[:loc] + snippet + file[loc:]
		err = ioutil.WriteFile(path, []byte(newFile), 0644)
		check(err)
	}
	return nil
}
func mapDispatchGenerator(file string, path string, actionPath string, config configuration) error {
	isMapDispatch := strings.Contains(file, config.mapDispatchMarker)
	if !isMapDispatch {
		regex := regexp.MustCompile("")
		newFile := regex.ReplaceAll([]byte(file), []byte(`const mapDispatchToProps = {
			...actionCreators,
		};
		`))
		error := ioutil.WriteFile(path, newFile, 0644)
		return error
	}
	return nil
}
func blankMapStateGenerator(file string, path string, actionPath string, config configuration) error {
	isMapState := strings.Contains(file, config.mapStateMarker)
	if !isMapState {
		regex := regexp.MustCompile("")
		newFile := regex.ReplaceAll([]byte(file), []byte(`const mapStateToProps = (state) => ({
		});
		`))
		error := ioutil.WriteFile(path, newFile, 0644)
		return error
	}
	return nil
}
func reducerGenerator(path string, reducerPath string, generateReducerFile bool, isReducerExist bool, config configuration, types, APItypes, actions []string) error {
	// Check if any existing reducers in the reducers file and then execute logic
	if generateReducerFile {

	} else if isReducerExist {
		file, err := ioutil.ReadFile(path)
		check(err)
		fmt.Println(file)
	}
	return nil
}
func combinedReducerGenerator(path string, reducerPath string, config configuration) error {
	// Check if combined reducers exists in the reducers file and then execute logic
	return nil
}

type configuration struct {
	spacer              string
	actionCreatorMarker string
	reactMarker         string
	flowMarker          string
	stateMarker         string
	stateTypesMarker    string
	propTypesMarker     string
	defaultPropsMarker  string
	reduxMarker         string
	mapDispatchMarker   string
	mapStateMarker      string
	actionPrefix        string
	actionSuffix        string
	reducerPrefix       string
	reducerSuffix       string
}

func main() {
	dir, err := os.Getwd()
	check(err)

	fmt.Println(dir)

	searchDir := dir

	generatePropTypes := false
	generateStateTypes := false
	generateMapDispatch := false
	generateBlankMapState := false

	generateReducers := false
	generateReducerFile := false
	generateCombinedReducer := false

	config := configuration{
		actionCreatorMarker: "actionCreators =",
		reactMarker:         "import React from ",
		flowMarker:          "@flow",
		stateMarker:         "state =",
		stateTypesMarker:    "State:",
		propTypesMarker:     "Props:",
		defaultPropsMarker:  "defaultProps:",
		reduxMarker:         "import {connect} from ",
		mapDispatchMarker:   "mapDispatchToProps",
		mapStateMarker:      "mapStateToProps",
		actionPrefix:        "",
		actionSuffix:        "Actions",
		reducerPrefix:       "",
		reducerSuffix:       "Reducer",
	}
	// import markers etc

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		check(err)
		if info.IsDir() || (!strings.Contains(path, ".js") && !strings.Contains(path, ".jsx")) {
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
		if isActionExist {
			types, APItypes, actions, err := parseActions(path, config)
			check(err)
			_, err = os.Stat(ReducerFile)
			isReducerExist := (err == nil)
			if generateReducers {
				reducerGenerator(path, ReducerFile, generateReducerFile, isReducerExist, config, types, APItypes, actions)
			}
			_, err = os.Stat(ReducerFile)
			isReducerExist = (err == nil)
			if generateCombinedReducer && isReducerExist {
				combinedReducerGenerator(path, ReducerFile, config)
			}
		}

		// append to files changed which should be returned
		fileList = append(fileList, path)
		return nil
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}
}
