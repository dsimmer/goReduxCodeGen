// @flow
import { connect } from 'react-redux';
import React from 'react';
import PropTypes from 'prop-types';
import {selectPage} from './../../DynamicReact/Actions'

class ShoppingCart extends React.Component<Props> {
    static defaultProps = {
        userDetails: {
            email: 'dylan',
            firstName: [
                "tester"
            ],
            lastName: 'lastName',
            mobileNumber: '11111111111',
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

export default connect(
	mapStateToProps,
	mapDispatchToProps,
)(ShoppingCart);
