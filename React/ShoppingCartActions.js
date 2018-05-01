// @flow
export const types = {
    ADD_TO_CART: 'ADD_TO_CART',
    REMOVE_FROM_CART: 'REMOVE_FROM_CART',
    EMPTY_CART: 'EMPTY_CART',
};

export const actionCreators = {
    addToCart: (items: any) => ({
        type: types.ADD_TO_CART,
        payload: items,
    }),
    removeFromCart: (items: any) => ({
        type: types.REMOVE_FROM_CART,
        payload: items,
    }),
    emptyCart: (items: any) => ({
        type: types.EMPTY_CART,
        payload: items,
    }),
};
