// @flow
import {types} from './ShoppingCartActions';

const cartIS = [];

/*
cart item:

{
    id: number
    count: number,
}
*/

export default function ShoppingCartReducer(state = cartIS, action) {
    if (action.type === types.ADD_TO_CART) {
        const present = state.findIndex((item) => item.id === action.payload.id);
        const newState = state.slice()
        if (present != -1) {
            newState[present].count += 1;
            return newState;
        } else {
            newState.push(action.payload);
            return newState;
        }
    }
    if (action.type === types.REMOVE_FROM_CART) {
        const present = state.findIndex((item) => item.id === action.payload.id);
        const newState = state.slice()
        if (present != -1) {
            if (newState[present].count > 1) {
                newState[present].count -= 1;
                return newState;
            } else {
                newState.splice(1, present);
                return newState;
            }
        } else {
            return state;
        }
    }
    if (action.type === types.EMPTY_CART) {
        return [];
    }
    return state;
}
