#Go Redux CodeGen

This small project will walk the working directory file tree and parse the js files it finds there.

To use this in its default state, your react components will have to be setup like the example included. No mapping of state or dispatch, no actions and no reducers. Just fill out your types (the component itself isnt necessary, but if you include it and the state and defaultprops we will generate flow types for them) then run the tool.

The tool will automagically fill out all that boilerplate in your corresponding files.

Your components will have to be using an isolated project structure like this one as currently I cannot map the files together if they are all in different directories. 

The configuration object is loaded from your config.json. It contains:

		connectorMarker:     "export default connect", // Identifier for an existing map configuration for your component
		spacer:              "1space" or "tab", // I haven't built an auto space detector yet, so you specify this manually (Xspace where x is a number or tab)
		actionCreatorMarker: "export const actionCreators = {", // Identifier for your action creator object. I use objects to keep my action creators more readable and easily importable
		reactMarker:         "import React from ", // Identifier for your component file so we know it is definitely react
		flowMarker:          "@flow", // Identifier for your component file so we know it is definitely using flow, otherwise we wont put in the flow types
		stateMarker:         "state =", // State marker for your component. You might assign an object in the constructor.
		defaultPropsMarker:  "static defaultProps = {", // Default props market, will turn this object into its flow definitions
		reduxMarker:         "react-redux", // Lets us know that you are using redux in this js file
		mapDispatchMarker:   "mapDispatchToProps", // So we dont double up in files that already map dispatch
		mapStateMarker:      "mapStateToProps", // So we dont double up in files that already map state
		actionPrefix:        "", // File naming config so we can associate your component file to its action and reducer files
		actionSuffix:        "Actions",
		reducerPrefix:       "",
		reducerSuffix:       "Reducers",
		fileExtension:       ".js", // the extension you use for your react files, I just use js

additionally you can pass flags to turn on/off particular generation in the config file
	generatePropTypes := true
	generateStateTypes := true
	generateMapDispatch := true
	generateBlankMapState := true
	generateConnector := true

	generateReducers := false
	generateReducerFile := false
	generateCombinedReducer := false

You can also configure the text that is inserted using the provided files

Happy generating!

