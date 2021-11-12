/**
 * Code.jsx
 * Author: 190010714
 * 
 * This file stores the info for rendering the Code page of our Journal
 */

import React from "react";
import Prism from "prismjs";
import CommentModal from "./CommentModal";

// import "./prism.css";

import {Helmet} from "react-helmet";



class Code extends React.Component {
    constructor(props){
        super(props)

        this.state = {
            file: window.projectName,
            project: window.projectID,
            content: 'hello',
            authorID: '11d38ba6c5-435b-11ec-bb68-320e0198aa16'

            
        }

        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleSubmit(e) {

    }
  componentDidMount() {
    
    // You can call the Prism.js API here
    setTimeout(() => Prism.highlightAll(), 0)
    console.log("ID Passed" + window.projectID);
    console.log("Project Name Passed" + window.projectName)
    
    

    let userID = null;                          //Preparing to get userID from session cookie
    let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
    for(let i = 0; i < cookies.length; i++){    //For each cookie,
        let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
        if(cookie[0].trim() == "userID"){       //  If userID key exists, extract the userID value
            userID = cookie[1].trim();
            break;
        }
    }

    if(userID === null){                        //If user has not logged in, disallow submit
        console.log("Not logged in");
        return;
    }

    this.props.code(this.state.file, this.state.project).then((files) => {
        console.log("received:" + files);
    }, (error) => {
        console.log(error);
    });

    // this.props.code(this.state.file, this.state.project, this.state.authorID, this.state.content).then((files) => {
    //     console.log("received:" + files);
    // }, (error) => {
    //     console.log(error);
    // });
   
    
    console.log("Code submitted");

  }
  render() {
    const code = ''
    return (
    <div className="renderCode">
        <h2 style={{textAlign: 'center',}} >Circular Queue</h2>
        <h4 style={{textAlign: 'center',}}>Author:  </h4>
        <CommentModal/>
        <pre className="line-numbers">
            <code className="language-java">
            {code}
            </code>
        </pre>
    </div>
    )
  }
}

export default Code;