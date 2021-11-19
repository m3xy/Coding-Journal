/**
 * Code.jsx
 * Author: 190010714
 * 
 * This file stores the info for rendering the Code page of our Journal
 */

import React from "react";
import Prism from "prismjs";
// import CommentModal from "./CommentModal";
import axiosInstance from "../Web/axiosInstance";

// import "./prism.css";
import { Button } from "react-bootstrap"
// import {Helmet} from "react-helmet";

const codeEndpoint = 'project/file'

class Code extends React.Component {
    constructor(props){
        super(props)

        this.state = {
            file: 'CountToFifteen.java',
            project: 8,
            content: 'hello',
            authorID: '11d38ba6c5-435b-11ec-bb68-320e0198aa16'

            
        }

        this.redirectToComment = this.redirectToComment.bind(this);
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

        //return new Promise((resolve, reject) => {

            // create a new XMLHttpRequest
        //    var request = new XMLHttpRequest()
            // get a callback when the server responds
        //    request.addEventListener('load', () => {
                // update the state of the component with the result here
        //        resolve(request.responseText);

                // TEMP: return response here, set the state of the login widget to be login approved
        //    })
            // open the request with the verb and the url TEMP: this will potentially change with actual URL
        //    request.open('POST', BACKEND_ADDRESS + codeEndpoint)
        //    request.onerror = reject;
        //    console.log(data);
        //    this.sendSecureRequest(request, data)
        // })
        axiosInstance.post(codeEndpoint, data)
                     .then(() => {console.log("received: " + files)})
                     .catch((error) => {console.log(error)});

    }

    redirectToComment() {
        var commentPage = window.open('/commentModal')
        commentPage.project = this.project 
        commentPage.file = this.file
        console.log('project: ' + this.project)
        console.log('file: ' + this.file)

    }

    componentDidMount() {
        // You can call the Prism.js API here
        setTimeout(() => Prism.highlightAll(), 0)
        console.log(window.project);

        let userId = null;                          //Preparing to get userId from session cookie
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userId"){       //  If userId key exists, extract the userId value
                userId = cookie[1].trim();
                break;
            }
        }

        if(userId === null){                        //If user has not logged in, disallow submit
            console.log("Not logged in");
            return;
        }

        // this.props.code(this.state.file, this.state.project).then((files) => {
        //    console.log("received:" + files);
        // }, (error) => {
        //    console.log(error);
        //});
        this.getCode(this.state.file, this.state.project);


        console.log("Code submitted");
    }

    render() {
        // const code = ''
        // console.log('Type: ' + typeof window.project)
        return (
        <div className="renderCode">
            <h2 style={{textAlign: 'center',}} >Circular Queue</h2>
            <h4 style={{textAlign: 'center',}}>Author:  </h4>
            <Button color='primary' className='px-4' onClick={this.redirectToComment}>
                Go To Comments
            </Button>
            <pre className="line-numbers">
                <code className="language-java">
                {window.project.content}
                </code>
            </pre>
        </div>
        )
    }
}

export default Code;
