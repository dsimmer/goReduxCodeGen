// @flow
import React from 'react';
import {selectPage} from './ShoppingCartActions';

class ShoppingCart extends React.Component<Props> {
    static defaultProps = {
        userDetails: {
            email: 'dylan',
            firstName: [
                "tester"
            ],
            lastName: 'lastName',
            mobileNumber: 0449637688,
        },
    }
    state = {
        suserDetails: {
            state_email: 'dylan',
            state_firstName: [
                "tester"
            ],
            state_lastName: 'lastName',
            state_mobileNumber: 0449637688,
        },
    }
	render() {
		return (
            <div>
                <i className="fa fa-shopping-cart" onClick={() => this.props.selectPage('checkout')} />
            </div>
		);
	}
}
