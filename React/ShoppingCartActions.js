// @flow
export const types = {
    ADD_TO_CART: 'ADD_TO_CART',
    REMOVE_FROM_CART: 'REMOVE_FROM_CART',
    EMPTY_CART: 'EMPTY_CART',
};

export const actionCreators = {
    addToCart: (item) => ({
        type: types.ADD_TO_CART,
        payload: item,
    }),
    removeFromCart: (item) => ({
        type: types.REMOVE_FROM_CART,
        payload: item,
    }),
    emptyCart: () => ({
        type: types.EMPTY_CART,
        payload: {},
    }),
};
