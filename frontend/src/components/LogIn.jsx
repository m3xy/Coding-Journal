/**
 * LogIn.jsx
 * author: 190010425
 * created: October 25, 2021
 * 
 * This file contains the login component and is in charge of sending the data
 * to the backend
 */

/**
 * TEMP:
 *  > add link to registration tab via "not a registered user? register here" link
 *  > add validation functionality for email and password fields
 */

import React from 'react';
import { Formik, Field, Form } from "formik";
import { DataWriter } from '../util' // TEMP: change this to not be a relative import

// constants
const HEADING = 'Login to Journal 11' // title to be displayed above login form
const EMAIL_MAX_LENGTH = 100 // maximum email length in number of chars
const PASSWORD_MAX_LENGTH = 64 // maximum password length in number of chars

/**
 * Web page for logging into the website. Uses Formik forms for the login interface
 */
class Login extends React.Component {
    constructor(props) {
        super(props)
        this.writer = new DataWriter() // object to write to the backend
        this.state = {
            loginSuccess: false, // to record whether a login is successful or not TEMP: set with the Datawriter
        }
    }

    /**
     * renders a Formik form for logging in. Formik takes care of password and
     * email constraints
     */
    render() {
        return (
            <div>
                <h1>{HEADING}</h1>
                <Formik
                    initialValues={{ email: "", password: "" }}
                    onSubmit={(values) => {
                        this.writer.sendLogIn(String(values.email), String(values.password))
                    }}
                >
                    <Form>
                        <Field
                            // validation is implicit for email type 
                            name="email"
                            type="email" 
                            placeholder="user@example.com"
                            validate={(values) => {this.validateEmail(String(values.email))}}
                        />
                        <Field
                            name="password" 
                            type="password" 
                            placeholder="Enter Password"
                            // extra validation for our own password requirements
                            validate={(values) => {this.validatePassword(String(values.password))}} 
                        />
                        <button type="submit">
                            Sign In
                        </button>
                    </Form>
                </Formik>
            </div>
        );
    }

    /**
     * validates that the user's email is valid
     * 
     * @param email the user email as a string
     * @return null if the email is valid, an error message as a string otherwise
     */
    validateEmail(email) {
        return (email.length > EMAIL_MAX_LENGTH) ? 'Email is over ' + String(EMAIL_MAX_LENGTH) + ' chars' : null
        // TEMP: add more validate conditions here
    }

    /**
     * validates that the user's email is valid
     * 
     * @param password the user's password as a string
     * @return null if the password is valid, an error message as a string otherwise
     */
    validatePassword(password) {
        return (password.length > PASSWORD_MAX_LENGTH) ? 'Password is over ' + PASSWORD_MAX_LENGTH + ' chars' : null
        // TEMP: add more validate conditions here
    }
}

export default Login;