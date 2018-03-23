// An auto dispatch wrapper for all actions, if they require values passed in they are auto passed. Passed into props with the same name.
const mapDispatchToProps = (dispatch) => {
    const actions = {};
    Object.entries(actionCreators).forEach((item) => {
        actions[item[0]] = (...params) => dispatch(item[1](...params));
    });
};

const mapDispatchToProps = {
    ...actionCreators,
};


// An auto reducer wrapper for all actions, defaults to a reducer that simply sets to the payload. Uses naming convention to figure out APIs and adds a loading reducer with a Loading suffix, that is all

// API calls are marked by having one each of _REQUEST, _ERROR and _REPLY
export default CombinedReducer = (() => {
    const ApiCalls = [];
    const typeMatched = {};
    types.forEach((type) => {
        if (type.slice(-8) === '_REQUEST') {
            const sanitized = type.slice(0, -8).replace('_', '').toLowerCase();
            ApiCalls.push(sanitized);
            typeMatched[sanitized] = type.slice(-8);
        } else if (type.slice(-6) !== '_REPLY' && type.slice(-6) !== '_ERROR') {
            const sanitized = type.replace('_', '').toLowerCase();
            typeMatched[sanitized] = type;
        }
    });
    const reducers = {};
    Object.keys(actionCreators).forEach((key) => {
        if (ApiCalls.includes(key.toLowerCase())) {
            reducers[key] = (state, action) => {
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_REQUEST`) {
                    return {};
                }
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_REPLY`) {
                    return action.payload;
                }
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_ERROR`) {
                    return action.payload;
                }
                return state;
            }
            reducers[`${key}Loading`] = (state, action) => {
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_REQUEST`) {
                    return true;
                }
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_REPLY`) {
                    return false;
                }
                if (action.type === `${type[typeMatched[key.toLowerCase()]]}_ERROR`) {
                    return false;
                }
                return state;
            }
        } else {
            reducers[key] = (state, action) => {
                if (action.type === type[typeMatched[key.toLowerCase()]]) {
                    return action.payload;
                }
                return state;
            }
        }
        // actions[item[0]] = () => dispatch(item[1](...params));
    });
    return (state, action) => {
        const newState = {};
        Object.keys(reducers).forEach((key) => {
            newState[key] = reducers[key](state[key], action);
        });
    }
})();
