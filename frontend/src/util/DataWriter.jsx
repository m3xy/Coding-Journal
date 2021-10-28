/**
 * DataWriter.jsx
 * author: 190010425
 * created: October 26, 2021
 * 
 * This file contains a class to send data to the Go backend code via http
 * All files send data through this class so that there is a common way of doing so
 */

import { BACKEND_INFO } from '../constants'

/**
 * Utility class with methods to send data to the backend via HTTP request
 * This class could be optimized by compiling many requests into a single
 * queue of requests.
 */
class DataWriter {
    /**
     * Method to send login credentials to the backend to be verified
     * 
     * @param userEmail the user's email. Each email can only have one account
     * @param userPassword the user's password as a string
     */
    sendLogIn(userEmail, userPassword) {
        // constructs JSON data to send to the backend
        let data = {
            email: userEmail,
            password: userPassword
        };

        // create a new XMLHttpRequest
        var request = new XMLHttpRequest()

        // get a callback when the server responds
        request.addEventListener('load', () => {
            // update the state of the component with the result here
            console.log(request.responseText)

            // TEMP: return response here, set the state of the login widget to be login approved
        })
        request.setRequestHeader() // TEMP: to set headers if necessary

        // open the request with the verb and the url TEMP: this will potentially change with actual URL
        request.open('POST', BACKEND_INFO.host + ':' + BACKEND_INFO.port)

        // send the request with the JSON data as body
        request.send(JSON.stringify(data))
    }

    // add more writing functions here
    // TEMP: return bool here to indicate success or failure
}

export default DataWriter;