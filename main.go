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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// SnakeToCamel converts our typenames to camel case
func SnakeToCamel(snake string) (camel string) {
	processed := strings.Split(snake, "_")
	for _, line := range processed {
		camel = camel + strings.Title(strings.ToLower(line))
	}
	return camel
}

type configuration struct {
	ConnectorMarker         string
	Spacer                  string
	ActionCreatorMarker     string
	ReactMarker             string
	FlowMarker              string
	StateMarker             string
	StateTypesMarker        string
	PropTypesMarker         string
	DefaultPropsMarker      string
	ReduxMarker             string
	MapDispatchMarker       string
	MapStateMarker          string
	ActionPrefix            string
	ActionSuffix            string
	ReducerPrefix           string
	ReducerSuffix           string
	FileExtension           string
	GeneratePropTypes       bool
	GenerateStateTypes      bool
	GenerateMapDispatch     bool
	GenerateBlankMapState   bool
	GenerateConnector       bool
	GenerateReducers        bool
	GenerateReducerFile     bool
	GenerateCombinedReducer bool
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

// Regex for identifying what kind of type a javascript variable is
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

// Some standard markers for finding spots in a JS file
var javascriptDict = map[string]string{
	// First empty line after all imports
	"afterImport": `(?:import.*?)([\r\n])`,
	"afterFlow":   `(@flow.*?)`,
}

// maybe just operate on an index?

// parseObjectValues takes in an object and uses the replacer function (replaceTypes is used exclusively here) to turn it into a flow definition
// pass a custom function to transform the object in different ways, it could for example generate TypeScript defs instead
func (fp *FileParser) parseObjectValues(file string, firstMarker string, lastMarker string, replacer func(string, bool) string) (string, error) {
	processed := file
	if !strings.Contains(processed, firstMarker) {
		return "", errors.New("No firstMarker found")
	}
	processedSplit := strings.SplitAfter(processed, firstMarker)

	processedFirst := strings.SplitAfter(processedSplit[1], lastMarker)
	processedLast := strings.SplitAfter(processedFirst[0], "\n")

	for index, line := range processedLast {
		if line == (fp.Config.Spacer + lastMarker) {
			fmt.Println("Ended on :" + line + "Due to: " + fp.Config.Spacer + lastMarker)
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

// parseActions takes in the actionsfile and finds and categorizes all the actions inside and sticks them on the fp struct
func (fp *FileParser) parseActions() ([]string, []string, []string, error) {
	var actions []string
	var types []string
	var APITypes []string

	processed := string(fp.actionFile)

	if !strings.Contains(processed, fp.Config.ActionCreatorMarker) {
		return types, APITypes, actions, errors.New("No actionCreator found")
	}
	processedSplit := strings.SplitAfter(processed, fp.Config.ActionCreatorMarker)
	// identifier is every type line until }
	processedTypes := strings.Split(processedSplit[0], "const types = {")
	if !(len(processedTypes) > 1) {
		return nil, nil, nil, errors.New("No types")
	}
	processedTypes = strings.Split(processedTypes[1], "\n")

	for _, line := range processedTypes {
		if strings.Contains(line, "}") {
			break
		} else if strings.Contains(line, "_REQUEST") {
			APITypes = append(APITypes, strings.Split(strings.Split(line, "'")[1], "_REQUEST")[0])
		} else if strings.Contains(line, "_REPLY") || strings.Contains(line, "_ERROR") {
			continue
		} else if strings.Contains(line, "'") {
			types = append(types, strings.Split(line, "'")[1])
		}
	}

	// identifier is any line prefixed by 4 spaces (setting and no more)
	// then get name

	processedAC := strings.SplitAfter(processedSplit[1], "\n")
	for _, line := range processedAC {
		regex := regexp.MustCompile("$(" + fp.Config.Spacer + "){1}")
		processLine := regex.FindIndex([]byte(line))
		if (processLine != nil) && (line != "};\n") {
			actions = append(actions, strings.SplitAfter(strings.SplitAfter(line, fp.Config.Spacer)[1], ":")[0])
		}
	}
	return types, APITypes, actions, nil
}

// replaceTypes simply identifies the flow types for the given value of state or props. Used in parseObjectValues to parse an object into flow type defs
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

// propTypesGenerator simply adds the flow types for props in if its missing from the component file
func (fp *FileParser) propTypesGenerator() error {
	isPropTypes := strings.Contains(fp.contents, fp.Config.PropTypesMarker)
	if !isPropTypes {
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		snippet, err := fp.parseObjectValues(fp.contents, fp.Config.DefaultPropsMarker, "}", replaceTypes)
		snippet = "\n\ntype PropTypes: {" + snippet + "\n};\n"
		check(err)
		fp.contents = fp.contents[:index] + snippet + fp.contents[index:]
	}
	return nil
}

// stateTypesGenerator simply adds the flow types for state in if its missing from the component file
func (fp *FileParser) stateTypesGenerator() error {
	isStateTypes := strings.Contains(fp.contents, fp.Config.StateTypesMarker)
	if !isStateTypes {
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		snippet, err := fp.parseObjectValues(fp.contents, fp.Config.StateMarker, "}", replaceTypes)
		snippet = "\ntype StateTypes: {" + snippet + "\n};\n"
		check(err)
		fp.contents = fp.contents[:index] + snippet + fp.contents[index:]
	}
	return nil
}

// mapDispatchGenerator simply adds the react redux mapDispatchToProps in if its missing from the component file
func (fp *FileParser) mapDispatchGenerator() error {
	isMapDispatch := strings.Contains(fp.contents, fp.Config.MapDispatchMarker)
	if !isMapDispatch {
		fp.contents = fp.contents + "\nconst mapDispatchToProps = {\n" + fp.Config.Spacer + "...actionCreators,\n};\n"
	}
	return nil
}

// blankMapStateGenerator simply adds the react redux mapStateToProps in if its missing from the component file
func (fp *FileParser) blankMapStateGenerator() error {
	isMapState := strings.Contains(fp.contents, fp.Config.MapStateMarker)
	if !isMapState {
		fp.contents = fp.contents + "\nconst mapStateToProps = (state) => ({\n});\n"
	}
	return nil
}

// connectorGenerator simply adds the react redux connector in if its missing from the component file
func (fp *FileParser) connectorGenerator() error {
	isConnector := strings.Contains(fp.contents, fp.Config.ConnectorMarker)
	if !isConnector {
		regex := regexp.MustCompile(javascriptDict["afterFlow"])
		loc := regex.FindAllIndex([]byte(fp.contents), -1)
		index := loc[len(loc)-1][1]
		fp.contents = fp.contents[:index] + "\nimport {connect} from 'react-redux';" + fp.contents[index:] + "\nexport default connect(\n" + fp.Config.Spacer + "mapStateToProps,\n" + fp.Config.Spacer + "mapDispatchToProps,\n)(ComponentName);\n"
	}
	return nil
}

// reducerGenerator augments an existing Reducer file with more functions (not adding to combined reducers) or generates the file from scratch
func (fp *FileParser) reducerGenerator(isReducerExist bool, types, APItypes, actions []string, actionFileName string) error {
	// Check if any existing reducers in the reducers file and then execute logic
	if fp.Config.GenerateReducerFile {
		// Generate reducer file from scratch and save

		newFile := `import {types} from './` + actionFileName + `';

`
		reducer := `
export default function combinedReducer(state, action) {
	return {`
		for _, name := range types {
			newFile = newFile + `export function ` + SnakeToCamel(name) + `Reducer(state, action) {
	if (action.type === types.` + name + `) {
		return action.payload
	}
	return state;
}
`
			reducer += "\n" + fp.Config.Spacer + fp.Config.Spacer + SnakeToCamel(name) + ": " + SnakeToCamel(name) + `Reducer(state, action),`
		}
		for _, name := range APItypes {
			newFile = newFile + `export function ` + name + `Reducer(state, action) {
	if (action.type === types.` + name + `_REPLY) {
		return action.payload
	}
	if (action.type === types.` + name + `_ERROR) {
		return action.payload
	}
	return state;
}
`
			newFile = newFile + `export function ` + SnakeToCamel(name) + `LoadingReducer(state, action) {
	if (action.type === types.` + name + `_REQUEST) {
		return true;
	}
	if (action.type === types.` + name + `_REPLY) {
		return false;
	}
	if (action.type === types.` + name + `_ERROR) {
		return false;
	}
	return state;
}
`
			reducer = "\n" + reducer + SnakeToCamel(name) + ": " + SnakeToCamel(name) + `Reducer(state, action),
		`
			reducer = reducer + name + "Loading: " + SnakeToCamel(name) + `LoadingReducer(state, action),`
		}

		reducer += `
	};
}
`
		fp.reducerFile = newFile + reducer
	} else if isReducerExist {
		// Add any reducer functions that dont exist
		// You have to manually update your combined reducer
		regex := regexp.MustCompile(javascriptDict["afterImport"])
		loc := regex.FindAllIndex([]byte(fp.reducerFile), -1)
		index := loc[len(loc)-1][1]
		var snippet string
		for _, name := range APItypes {
			if !strings.Contains(fp.reducerFile, name) {
				snippet = snippet + `
export function ` + SnakeToCamel(name) + `Reducer(state, action) {
	if (action.type === types.` + name + `_REPLY) {
		return action.payload
	}
	if (action.type === types.` + name + `_ERROR) {
		return action.payload
	}
	return state;
}
`
				snippet = snippet + `export function ` + SnakeToCamel(name) + `LoadingReducer(state, action) {
	if (action.type === types.` + name + `_REQUEST) {
		return true;
	}
	if (action.type === types.` + name + `_REPLY) {
		return false;
	}
	if (action.type === types.` + name + `_ERROR) {
		return false;
	}
	return state;
}
`
			}
		}
		for _, name := range types {
			if !strings.Contains(fp.reducerFile, name) {
				snippet = snippet + `export function ` + SnakeToCamel(name) + `Reducer(state, action) {
	if (action.type === types.` + name + `) {
		return action.payload
	}
	return state;
}
`
			}
		}

		fp.reducerFile = fp.reducerFile[:index] + snippet + fp.reducerFile[index:]
	}
	return nil
}

// combinedReducerGenerator is not used in v1
func (fp *FileParser) combinedReducerGenerator() error {
	// Check if combined reducers exists in the reducers file and then execute logic

	// Not implementing in v1 as it doesnt seem useful - generate the reducer file from scratch if you want this
	return nil
}

// ProcessFile uses the info from the FileParser struct to parse and generate the relevant js.
func (fp *FileParser) ProcessFile() error {
	if fp.Info.IsDir() || (!strings.Contains(fp.Path, ".js") && !strings.Contains(fp.Path, ".jsx")) {
		return nil
	}
	contents, err := ioutil.ReadFile(fp.Path)
	fp.contents = string(contents)
	check(err)

	isReact := strings.Contains(fp.contents, fp.Config.ReactMarker)
	isFlow := strings.Contains(fp.contents, fp.Config.FlowMarker)

	ActionFileName := fp.Config.ActionPrefix + strings.TrimSuffix(fp.Info.Name(), fp.Config.FileExtension) + fp.Config.ActionSuffix + fp.Config.FileExtension
	ActionPath := filepath.Dir(fp.Path) + string(os.PathSeparator) + ActionFileName
	_, err = os.Stat(ActionPath)
	isActionExist := err == nil
	if isActionExist {
		actionfile, err := ioutil.ReadFile(ActionPath)
		check(err)
		fp.actionFile = string(actionfile)
	}

	if isFlow && isReact {
		isDefaultProps := strings.Contains(fp.contents, fp.Config.DefaultPropsMarker)
		isState := strings.Contains(fp.contents, fp.Config.StateMarker)

		if fp.Config.GenerateStateTypes && isState {
			fp.stateTypesGenerator()
		}
		if fp.Config.GeneratePropTypes && isDefaultProps {
			fp.propTypesGenerator()
		}
	}
	if isReact && isActionExist {
		isMapState := strings.Contains(fp.contents, fp.Config.MapStateMarker)
		isConnector := strings.Contains(fp.contents, fp.Config.ConnectorMarker)

		if fp.Config.GenerateMapDispatch {
			fp.mapDispatchGenerator()
		}
		if fp.Config.GenerateBlankMapState && !isMapState {
			fp.blankMapStateGenerator()
		}
		if fp.Config.GenerateConnector && !isConnector {
			fp.connectorGenerator()
		}
	}
	ReducerFileName := fp.Config.ReducerPrefix + strings.TrimSuffix(fp.Info.Name(), fp.Config.FileExtension) + fp.Config.ReducerSuffix + fp.Config.FileExtension
	ReducerPath := filepath.Dir(fp.Path) + string(os.PathSeparator) + ReducerFileName
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
		if fp.Config.GenerateReducers {
			fp.reducerGenerator(isReducerExist, types, APItypes, actions, ActionFileName)
		}
		// Not in v1
		// if fp.Config.generateCombinedReducer && isReducerExist {
		// 	fp.combinedReducerGenerator()
		// }
	}
	if isReducerExist || (fp.Config.GenerateReducerFile && isActionExist) {
		err := ioutil.WriteFile(ReducerPath, []byte(fp.reducerFile), 0644)
		check(err)
	}
	if isActionExist {
		err = ioutil.WriteFile(ActionPath, []byte(fp.actionFile), 0644)
		check(err)
		err = ioutil.WriteFile(fp.Path, []byte(fp.contents), 0644)
		check(err)
	}

	return nil
}

// main simply walks the directory it is invoked in
func main() {
	searchDir, err := os.Getwd()
	check(err)

	configFile, err := ioutil.ReadFile("./config.json")
	check(err)
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
