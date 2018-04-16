package main

import (
	"encoding/json"
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

type configuration struct {
	connectorMarker         string
	spacer                  string
	actionCreatorMarker     string
	reactMarker             string
	flowMarker              string
	stateMarker             string
	stateTypesMarker        string
	propTypesMarker         string
	defaultPropsMarker      string
	reduxMarker             string
	mapDispatchMarker       string
	mapStateMarker          string
	actionPrefix            string
	actionSuffix            string
	reducerPrefix           string
	reducerSuffix           string
	fileExtension           string
	generatePropTypes       bool
	generateStateTypes      bool
	generateMapDispatch     bool
	generateBlankMapState   bool
	generateConnector       bool
	generateReducers        bool
	generateReducerFile     bool
	generateCombinedReducer bool
}

/*
FileParser holds the file to parse and the appropriate methods. Require for init are the public properties:
Config
Path
Info
*/
type FileParser struct {
	contents    string
	reducerFile string
	actionFile  string
	Config      configuration
	Path        string
	Info        os.FileInfo
}

var typeDict = map[string]string{
	"undefined": `\s*(undefined),?\s*`,
	"null":      `\s*(null),?\s*`,
	"float":     `^\s*[+-]?[0-9]*[\.][0-9]+,?\s*$`,
	"number":    `^\s*[+-]?[0-9]+,?\s*$`,
	"array":     `\s*\[\s*`,
	"arrayEnd":  `\s*\],?\s*`,
	"object":    `\s*[\{\}],?\s*`,
	"string":    `["'].*["'],?`,
}
var javascriptDict = map[string]string{
	// First empty line after all imports
	"afterImport": `(?:import.*?)([\r\n])`,
	"afterFlow":   `(@flow.*?)`,
}

// todo change all splits to regex - so we can get the closing brackets correctly
// TODO insert React.Component<Props> and State
// TODO: relies on types being before actioncreators - good practice anyways

func (fp FileParser) parseObject(file string, firstMarker string, lastMarker string, spacer string) ([]string, error) {
	var lines []string

	processed := file
	if !strings.Contains(processed, firstMarker) {
		return lines, errors.New("No firstMarker found")
	}
	processedSplit := strings.SplitAfter(processed, firstMarker)

	processedFirst := strings.SplitAfter(processedSplit[1], lastMarker)
	processedLast := strings.SplitAfter(processedFirst[0], "\n")

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

func (fp FileParser) parseObjectValues(file string, firstMarker string, lastMarker string, replacer func(string, bool) string) (string, error) {
	processed := file
	if !strings.Contains(processed, firstMarker) {
		return "", errors.New("No firstMarker found")
	}
	processedSplit := strings.SplitAfter(processed, firstMarker)

	processedFirst := strings.SplitAfter(processedSplit[1], lastMarker)
	processedLast := strings.SplitAfter(processedFirst[0], "\n")

	for index, line := range processedLast {
		if line == (fp.Config.spacer + lastMarker) {
			fmt.Println("Ended on :" + line + "Due to: " + fp.Config.spacer + lastMarker)
			break
		} else {
			value := strings.SplitAfter(line, ":")
			if len(value) > 1 {
				value[1] = replacer(value[1], true)
				processedLast[index] = strings.Join(value, "")
			} else {
				value[0] = replacer(value[0], false)
				processedLast[index] = strings.Join(value, "")
			}
		}
	}

	// join string up again
	newFile := strings.Join(processedLast, "")

	return newFile, nil
}

// TODO: relies on types being before actioncreators - good practice anyways
func (fp FileParser) parseActions() ([]string, []string, []string, error) {
	var actions []string
	var types []string
	var APITypes []string

	processed := string(fp.actionFile)

	if !strings.Contains(processed, fp.Config.actionCreatorMarker) {
		return types, APITypes, actions, errors.New("No actionCreator found")
	}
	processedSplit := strings.SplitAfter(processed, fp.Config.actionCreatorMarker)
	// identifier is every type line until }
	processedTypes := strings.SplitAfter(processedSplit[0], "types = {")
	processedTypes = strings.SplitAfter(processedTypes[1], "\n")

	for _, line := range processedTypes {
		if strings.Contains(line, "}") {
			break
		} else if strings.Contains(line, "_REQUEST") {
			APITypes = append(APITypes, strings.SplitAfter(strings.SplitAfter(line, "'")[1], "_REQUEST")[0])
		} else if strings.Contains(line, "_REPLY") || strings.Contains(line, "_ERROR") {
			continue
		} else if strings.Contains(line, "'") {
			types = append(types, strings.SplitAfter(line, "'")[1])
		}
	}

	// identifier is any line prefixed by 4 spaces (setting and no more)
	// then get name

	processedAC := strings.SplitAfter(processedSplit[1], "\n")
	for _, line := range processedAC {
		regex := regexp.MustCompile("$(" + fp.Config.spacer + "){1}")
		processLine := regex.FindIndex([]byte(line))
		if (processLine != nil) && (line != "};\n") {
			actions = append(actions, strings.SplitAfter(strings.SplitAfter(line, fp.Config.spacer)[1], ":")[0])
		}
	}

	return types, APITypes, actions, nil
}

func replaceTypes(inputString string, split bool) string {
	input := []byte(inputString)

	regexUndefined := regexp.MustCompile(typeDict["undefined"])
	regexNull := regexp.MustCompile(typeDict["null"])
	regexFloat := regexp.MustCompile(typeDict["float"])
	regexNumber := regexp.MustCompile(typeDict["number"])
	regexArray := regexp.MustCompile(typeDict["array"])
	regexArrayEnd := regexp.MustCompile(typeDict["arrayEnd"])
	regexObject := regexp.MustCompile(typeDict["object"])
	regexString := regexp.MustCompile(typeDict["string"])

	var regexResult [][]int
	end := ""
	if split {
		end = "\n"
	}
	if regexResult = regexUndefined.FindAllIndex(input, -1); regexResult != nil {
		return " undefined," + end
	} else if regexResult = regexNull.FindAllIndex(input, -1); regexResult != nil {
		return " null," + end
	} else if regexResult = regexFloat.FindAllIndex(input, -1); regexResult != nil {
		return " float," + end
	} else if regexResult = regexNumber.FindAllIndex(input, -1); regexResult != nil {
		return " number," + end
	} else if regexResult = regexArray.FindAllIndex(input, -1); regexResult != nil {
		return " Array<"
	} else if regexResult = regexArrayEnd.FindAllIndex(input, -1); regexResult != nil {
		return " >,\n"
	} else if regexResult = regexObject.FindAllIndex(input, -1); regexResult != nil {
		return inputString
	} else if regexResult = regexString.FindAllIndex(input, -1); regexResult != nil {
		return " string," + end
	}
	return inputString
}

func (fp FileParser) propTypesGenerator() error {
	isPropTypes := strings.Contains(fp.contents, fp.Config.propTypesMarker)
	if isPropTypes {
		// Not needed?
		// existingProps, err := parseObject(path, Config.propTypesMarker, "}", Config.spacer)
		// check(err)
		//insert
	} else {
		// create from scratch
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		snippet, err := fp.parseObjectValues(fp.contents, fp.Config.defaultPropsMarker, "}", replaceTypes)
		snippet = "\n\ntype PropTypes: {" + snippet + "\n};\n"
		check(err)
		fp.contents = fp.contents[:index] + snippet + fp.contents[index:]
	}
	return nil
}
func (fp FileParser) stateTypesGenerator() error {
	isStateTypes := strings.Contains(fp.contents, fp.Config.stateTypesMarker)
	if isStateTypes {
		// not needed?
		// existingStates, err := parseObject(path, Config.stateTypesMarker, "}", Config.spacer)
		// check(err)
		//insert
	} else {
		// create from scratch
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		snippet, err := fp.parseObjectValues(fp.contents, fp.Config.stateMarker, "}", replaceTypes)
		snippet = "\ntype StateTypes: {" + snippet + "\n};\n"
		check(err)
		fp.contents = fp.contents[:index] + snippet + fp.contents[index:]
	}
	return nil
}
func (fp FileParser) mapDispatchGenerator() error {
	isMapDispatch := strings.Contains(fp.contents, fp.Config.mapDispatchMarker)
	if !isMapDispatch {
		fp.contents = fp.contents + "\nconst mapDispatchToProps = {\n...actionCreators,\n};\n"
	}
	return nil
}
func (fp FileParser) blankMapStateGenerator() error {
	isMapState := strings.Contains(fp.contents, fp.Config.mapStateMarker)
	if !isMapState {
		fp.contents = fp.contents + "\nconst mapStateToProps = (state) => ({\n});\n"
	}
	return nil
}
func (fp FileParser) connectorGenerator() error {
	isConnector := strings.Contains(fp.contents, fp.Config.connectorMarker)
	if !isConnector {
		regex := regexp.MustCompile(javascriptDict["afterFlow"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		fp.contents = fp.contents[:index] + "\nimport {connect} from 'react-redux';" + fp.contents[index:] + "\nexport default connect(\n" + fp.Config.spacer + "mapStateToProps,\n" + fp.Config.spacer + "mapDispatchToProps,\n)(ComponentName);\n"
	}
	return nil
}
func (fp FileParser) reducerGenerator(isReducerExist bool, types, APItypes, actions []string) error {
	// Check if any existing reducers in the reducers file and then execute logic
	if fp.Config.generateReducerFile {

	} else if isReducerExist {
		fmt.Println("check")
	}
	return nil
}
func (fp FileParser) combinedReducerGenerator() error {
	// Check if combined reducers exists in the reducers file and then execute logic
	return nil
}

// ProcessFile uses the info from the FileParser struct to parse and generate the relevant js.
func (fp FileParser) ProcessFile() error {
	if fp.Info.IsDir() || (!strings.Contains(fp.Path, ".js") && !strings.Contains(fp.Path, ".jsx")) {
		return nil
	}
	contents, err := ioutil.ReadFile(fp.Path)
	fp.contents = string(contents)
	check(err)

	isReact := strings.Contains(fp.contents, fp.Config.reactMarker)
	isFlow := strings.Contains(fp.contents, fp.Config.flowMarker)
	// isRedux := strings.Contains(fp.contents, fp.Config.reduxMarker)

	ActionPath := filepath.Dir(fp.Path) + string(os.PathSeparator) + fp.Config.actionPrefix + strings.TrimSuffix(fp.info.Name(), fp.Config.fileExtension) + fp.Config.actionSuffix + fp.Config.fileExtension
	_, err = os.Stat(ActionPath)
	isActionExist := err == nil
	if isActionExist {
		actionfile, err := ioutil.ReadFile(ActionPath)
		check(err)
		fp.actionFile = string(actionfile)
	}

	if isFlow && isReact {
		isDefaultProps := strings.Contains(fp.contents, fp.Config.defaultPropsMarker)
		isState := strings.Contains(fp.contents, fp.Config.stateMarker)

		if fp.Config.generateStateTypes && isState {
			fp.stateTypesGenerator()
		}
		if fp.Config.generatePropTypes && isDefaultProps {
			fp.propTypesGenerator()
		}
	}
	if isReact && isActionExist {
		isMapState := strings.Contains(fp.contents, fp.Config.mapStateMarker)
		isConnector := strings.Contains(fp.contents, fp.Config.connectorMarker)

		if fp.Config.generateMapDispatch {
			fp.mapDispatchGenerator()
		}
		if fp.Config.generateBlankMapState && !isMapState {
			fp.blankMapStateGenerator()
		}
		if fp.Config.generateConnector && !isConnector {
			fp.connectorGenerator()
		}
	}
	ReducerPath := filepath.Dir(fp.Path) + string(os.PathSeparator) + fp.Config.reducerPrefix + strings.TrimSuffix(fp.Info.Name(), fp.Config.fileExtension) + fp.Config.reducerSuffix + fp.Config.fileExtension
	_, err = os.Stat(ReducerPath)
	isReducerExist := (err == nil)
	if isActionExist {
		types, APItypes, actions, err := fp.parseActions()
		check(err)
		if isReducerExist {
			reducerfile, err := ioutil.ReadFile(ReducerPath)
			check(err)
			fp.reducerFile = string(reducerfile)
		}
		if fp.Config.generateReducers {
			fp.reducerGenerator(isReducerExist, types, APItypes, actions)
		}
		_, err = os.Stat(ReducerPath)
		isReducerExist = (err == nil)
		if fp.Config.generateCombinedReducer && isReducerExist {
			fp.combinedReducerGenerator()
		}
	}
	if isReducerExist {
		err := ioutil.WriteFile(ReducerPath, []byte(fp.actionFile), 0644)
		check(err)
	}
	if isActionExist {
		err = ioutil.WriteFile(ActionPath, []byte(fp.reducerFile), 0644)
		check(err)
	}
	err = ioutil.WriteFile(fp.Path, []byte(fp.contents), 0644)
	check(err)

	return nil
}

func main() {
	searchDir, err := os.Getwd()
	check(err)

	configFile, err := ioutil.ReadFile("config.json")

	config := configuration{}
	err = json.Unmarshal(configFile, &config)
	check(err)
	// import markers etc
	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		check(err)
		newFileParser := FileParser{
			Path:   path,
			Info:   info,
			Config: config,
		}
		return newFileParser.ProcessFile()
	})

	check(err)
	fmt.Println("Done!")
}
