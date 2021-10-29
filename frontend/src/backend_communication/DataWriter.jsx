/**
 * DataWriter.jsx
 * author: 190010425
 * created: October 26, 2021
 * 
 * This file contains a class to send data to the Go backend code via http
 * All files send data through this class so that there is a common way of doing so
 */

/**
 * Utility class with methods to send data to the backend via HTTP request
 * This class could be optimized by compiling many requests into a single
 * queue of requests.
 * 
 * TEMP: could generalize all get and post requests to be sent by the same function
 */
class DataWriter {

    constructor(backend_host, backend_port) {
        this.backend_host = backend_host
        this.backend_port = backend_port
    }

    /**
     * Method to send login credentials to the backend to be verified
     * 
     * @param userEmail the user's email. Each email can only have one account
     * @param userPassword the user's password as a string
     */
    loginUser(userEmail, userPassword) {
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
        // request.setRequestHeader() // TEMP: to set headers if necessary

        // open the request with the verb and the url TEMP: this will potentially change with actual URL
        request.open('POST', this.backend_host + ':' + this.backend_port)

        // send the request with the JSON data as body
        request.send(JSON.stringify(data))

        // TEMP: return bool here to indicate success or failure
    }


    /**
     * Sends a POST request to the go server to register a new user
     * 
     * @param userFirstName the first name of the user being registered
     * @param userLastName the last name of the user being registered
     * @param userEmail the user's email (only one account per email)
     * @param userPassword the user's password
     */
    registerUser(userFirstName, userLastName, userEmail, userPassword) {
        // constructs JSON data to send to the backend
        let data = {
            firstname: userFirstName,
            lastname: userLastName,
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
        // request.setRequestHeader() // TEMP: to set headers if necessary

        // open the request with the verb and the url TEMP: this will potentially change with actual URL
        request.open('POST', this.backend_host + ':' + this.backend_port)

        // send the request with the JSON data as body
        request.send(JSON.stringify(data))
    }

    // add more writing functions here
}

export default DataWriter;