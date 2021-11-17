/**
 * DataWriter.jsx
 * author: 190010425, 190019931, 190010714
 * created: October 26, 2021
 * 
 * This file contains a class to send data to the Go backend code via http
 * All files send data through this class so that there is a common way of doing so
 */

import { history } from "../_helpers"

// URL endpoints for backend functions
const loginEndpoint = '/glogin'
const registerEndpoint = '/register'
const uploadEndpoint = '/upload'
const profileEndpoint = '/users'

const codeEndpoint = '/project/file'
const commentEndpoint = '/project/file/newcomment'

const BACKEND_TOKEN = process.env.BACKEND_TOKEN;
const BACKEND_ADDRESS = process.env.BACKEND_ADDRESS;

/**
 * Utility class with methods to send data to the backend via HTTP request
 * This class could be optimized by compiling many requests into a single
 * queue of requests.
 * 
 * TEMP: could generalize all get and post requests to be sent by the same function
 */
class DataWriter {

    constructor(backend_host, backend_port) {
        this.backend_host = backend_host;
        this.backend_port = backend_port;
        // this.userID = null;
    }

    /**
     * Author: 190014935
     * Send a request with needed security token and headers.
     *
     * @param request The request to send.
     * @param data The body for the request.
     */
    sendSecureRequest(request, data) {
        request.setRequestHeader("X-FOREIGNJOURNAL-SECURITY-TOKEN", BACKEND_TOKEN)
        if (data == null) {
            request.send()
        } else {
            request.send(JSON.stringify(data))
        }
    }

    /**
     * Method to send login credentials to the backend to be verified
     * 
     * @param userEmail the user's email. Each email can only have one account
     * @param userPassword the user's password as a string
     */
    loginUser(userEmail, userPassword, journal) {

        // constructs JSON data to send to the backend
        let data = {
            email: userEmail,
            password: userPassword,
            groupNumber:journal
        };

        // create a new XMLHttpRequest
        var request = new XMLHttpRequest()
        
        // var response;

        // get a callback when the server responds
        request.addEventListener('load', () => {
            // update the state of the component with the result here
            console.log(request.responseText)
            // this.userID = request.responseText;

            document.cookie = "userID=" + request.responseText + "; SameSite=Lax"; //Set a session cookie for logged in user
            history.push('/');

            // TEMP: return response here, set the state of the login widget to be login approved
        })
        // open the request and send request.
        request.open('POST', BACKEND_ADDRESS + loginEndpoint)
        this.sendSecureRequest(request, data)

        // Send request with JSON data as body.
        // TEMP: return bool here to indicate success or failure#
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
            console.log(request.responseText);
            if(request.status == 200){
                history.push('/');
            }

            // TEMP: return response here, set the state of the login widget to be login approved
        })
        // open the request with the verb and the url TEMP: this will potentially change with actual URL
        request.open('POST', BACKEND_ADDRESS + registerEndpoint)
        this.sendSecureRequest(request, data)
    }

     /**
     * Author: 190010714
     * Sends a POST request to the go server to register a new user
     * 
     * @param file the file ID for which we wish to render the code
     * @param project the project ID in which the file is located
     */
    getCode(file, project) {
        // constructs JSON data to send to the backend
        let data = {
            filePath: file,
            projectId: project
        };

        return new Promise((resolve, reject) => {

            // create a new XMLHttpRequest
            var request = new XMLHttpRequest()
            // get a callback when the server responds
            request.addEventListener('load', () => {
                // update the state of the component with the result here
                resolve(request.responseText);

                // TEMP: return response here, set the state of the login widget to be login approved
            })
            // open the request with the verb and the url TEMP: this will potentially change with actual URL
            request.open('POST', BACKEND_ADDRESS + codeEndpoint)
            request.onerror = reject;
            console.log(data);
            this.sendSecureRequest(request, data)
        })

    }
    
     /**
     * Author: 190010714
     * Sends a POST request to the go server to uplaod a new comment
     * 
     * @param file the file ID for the file on which the comment was made
     * @param project the project ID for the project in which the file is in
     * @param author the author of the comment
     * @param content the content of the comment
     */
    uploadComment(file, project, author, content) {
        // constructs JSON data to send to the backend
        let data = {
            filePath: file,
            projectId: project,
            author: author,
            content: content
        };

        return new Promise((resolve, reject) => {

            // create a new XMLHttpRequest
            var request = new XMLHttpRequest()
            // get a callback when the server responds
            request.addEventListener('load', () => {
                // update the state of the component with the result here
                resolve(request.responseText);

                // TEMP: return response here, set the state of the login widget to be login approved
            })
            // open the request with the verb and the url TEMP: this will potentially change with actual URL
            request.open('POST', BACKEND_ADDRESS + commentEndpoint)
            request.onerror = reject;
            console.log(data);
            this.sendSecureRequest(request, data)
        })

    }

    /**
     * Sends a POST request to the go server to upload (project) files
     * 
     * @param {JSON} userID Project files' Author's User ID
     * @param {Array.<File>} files Project files
     * @returns 
     */
    uploadFiles(userID, projectName, files) {

        if(userID === null) {
            console.log("not logged in!");
            return;
        }

        const authorID = JSON.parse(userID).userId;  //Extract author's userID

        const filePromises = files.map((file, i) => {   //Create Promise for each file (Encode to base 64 before upload)
            return new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.readAsDataURL(file);
                reader.onload = function(e) {
                    files[i] = e.target.result;
                    resolve();                          //Promise(s) resolved/fulfilled once reader has encoded file(s) into base 64
                } 
                reader.onerror = function() {
                    reject();
                }
            });
        })

        return new Promise ( (resolve, reject) => {

            var request = new XMLHttpRequest();

            Promise.all(filePromises) //Encode all files to upload into base 64 before uploading
                .then(() => {
                    let data = {
                        author : authorID,
                        name : projectName,
                        content : files[0]
                    }
    
                    console.log(data);
            
                    // get a callback when the server responds
                    request.addEventListener('load', () => {
                        //Return response for code page
                        resolve(request.responseText);
                    })

                    request.onerror = reject;
            
                    // open the request with the verb and the url TEMP: this will potentially change with actual URL
                    request.open('POST', BACKEND_ADDRESS + uploadEndpoint)
                    this.sendSecureRequest(request, data)
                })
                .catch((error) => {
                    console.log(error);
                })
        })
    }


    /**
     * Author: 190019931
     * Get a user's profile information.
     *
     * @param userID The user to fetch profile from.
     */
    getProfile(userID) {
        return new Promise((resolve, reject) => {

            // return resolve({                 //Testing
            //     firstname: "John",
            //     lastname: "Doe",
            //     email: "JohnDoe@gmail.com",
            //     usertype: 4,
            //     phonenumber: "012345678910",
            //     organization: "Lonely guy",
            //     projects : {
            //         1: "proj1",
            //         2 : "proj2"
            //     }
            // });

            var request = new XMLHttpRequest();
            
            request.open('GET', BACKEND_ADDRESS + profileEndpoint + "/" + userID);
            // get a callback when the server responds
            request.addEventListener('load', () => {
                resolve(request.responseText);
            })
            request.onerror = reject;
            this.sendSecureRequest(request, null)
        })
    }
    // add more writing functions here
}

export default DataWriter;
