/**
 * App.jsx
 * author: 190010714, 190010425
 * 
 * This file holds the main framework for the website.
 * 
 * To add a new page:
 *  (1) Code a new class/function to render a web page and add it to the Pages/ directory
 *  (2) Add a new entry in the Pages/index.js file to export your webpage and import it here
 *  (3) add a new <Route ... /> mapping a URL to your web page
 *  (4) add a new <Link /> component in the Pages/Navigation.jsx file
 */

import React from 'react';
import { Router, Route, Switch, Redirect } from 'react-router-dom';
import { connect } from 'react-redux';

import { history } from '../_helpers';
import { alertActions } from '../_actions';
import { PrivateRoute } from '../_components';

import { DataWriter } from '../backend_communication';
import { Navigation, Home, Login, Register, About, Contact, Footer } from '../Pages'

// defines website constants here for ease of configuration. 
// TEMP: could be moved to another file
// TEMP: change constants on integration with backend
const constants = {
    frontend: {
        host: 'http://localhost',
        port: '23409' // TEMP: for now
    },
    backend: {
        host: 'http://localhost',
        port: '3333'
    }
}

class App extends React.Component {
    constructor(props) {
        super(props);

        // data writing module for communicating with the backend
        this.writer = new DataWriter(constants.backend.host, constants.backend.port)

        history.listen((location, action) => {
            // clear alert on location change
            this.props.clearAlerts();
        });
    }

    render() {
        const { alert } = this.props;
        return (
            <div className="jumbotron">
                <div className="container">
                    <div className="col-sm-8 col-sm-offset-2">
                        {alert.message &&
                            <div className={`alert ${alert.type}`}>{alert.message}</div>
                        }
                        <Router history={history}>
                            <Navigation />
                            <Switch>
                                <PrivateRoute exact path="/" component={() => <Home />} />
                                <Route path="/login" component={() => <Login 
                                    login={(email, password) => this.writer.loginUser(email, password)}/>} 
                                />
                                <Route path="/register" component={() => <Register
                                    register={(firstname, lastname, email, password) => this.writer.registerUser(firstname, lastname, email, password)}/>} 
                                />
                                <Route path="/about" exact component={() => <About />} />
                                <Route path="/contact" exact component={() => <Contact />} />
                                <Redirect from="*" to="/" />
                            </Switch>
                            <Footer />
                        </Router>
                    </div>
                </div>
            </div>
        );
    }
}

function mapState(state) {
    const { alert } = state;
    return { alert };
}

const actionCreators = {
    clearAlerts: alertActions.clear
};

const connectedApp = connect(mapState, actionCreators)(App);
export { connectedApp as App };