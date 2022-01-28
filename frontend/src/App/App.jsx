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
import { BrowserRouter, Route, Routes/*, Navigate*/ } from 'react-router-dom';
import { connect } from 'react-redux';
// import { history } from '../_helpers';

import { alertActions } from '../_actions';
// import { PrivateRoute } from '../_components';

import { Navigation, CommentModal, Home, Login, Register, About, Contact, Footer, Code, Upload, Profile } from '../Pages'


// import 'bootstrap/dist/css/bootstrap.min.css'

// defines website constants here for ease of configuration. 
// TEMP: could be moved to another file
// TEMP: change constants on integration with backend
// const constants = {
    // frontend: {
    //    host: 'http://localhost',
    //    port: '23409' // TEMP: for now
    // },
    // backend: {
	//     host: 'http://localhost',
    //     port: '3333'
    // }
// }

class App extends React.Component {
    constructor(props) {
        super(props);
        // history.listen((/*location, action*/) => {
            // clear alert on location change
        //     this.props.clearAlerts();
        // });
    }

    componentDidMount() {
        // this.deleteCookies();
    }

    /**
     * Deletes all cookies
     */
    deleteCookies() {
        var cookies = document.cookie.split(';'); 
    
        // The "expire" attribute of every cookie is set to "Thu, 01 Jan 1970 00:00:00 GMT".
        for (var i = 0; i < cookies.length; i++) {
            document.cookie = cookies[i] + "=;expires=" + new Date(0).toUTCString();  //Setting all cookies expiry date to be a past date.
        }
    }

    render() {
        return (
            <BrowserRouter history={history} >
                <Navigation />
                <Routes>
                    <Route exact path="/" element = {<Home/>} />
                    <Route path="/login" element = {<Login />} />
                    <Route path="/register" element = {<Register />} />
                    <Route path="/about" element = {<About />} />
                    <Route path="/code" element = {<Code />} />
                    <Route path="/contact" element = {<Contact/>} />
                    <Route path="/commentModal" element = {<CommentModal />} / >
                    <Route path="/upload" element = {<Upload/>} />
                    <Route path="/profile" element = {<Profile/>} />
                    {/*<Route path="*" element={<Navigate to='/' replace />} />*/}
                </Routes>
                <Footer />
            </ BrowserRouter>
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
