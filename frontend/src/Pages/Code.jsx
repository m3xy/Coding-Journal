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
import { Button } from "react-bootstrap"
import {Helmet} from "react-helmet";



class Code extends React.Component {
    constructor(props){
        super(props)

        this.state = {
            file: 'CountToFifteen.java',
            submission: 8,
            content: 'hello',
            authorID: '11d38ba6c5-435b-11ec-bb68-320e0198aa16'

            
        }

        this.redirectToComment = this.redirectToComment.bind(this);
    }

redirectToComment() {
        var commentPage = window.open('/commentModal')
        commentPage.submission = this.submission 
        commentPage.file = this.file
        console.log('submission: ' + this.submission)
        console.log('file: ' + this.file)

    }
  componentDidMount() {
    
    // You can call the Prism.js API here
    setTimeout(() => Prism.highlightAll(), 0)
    console.log(window.submission);
   
    
    

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

    this.props.code(this.state.file, this.state.submission).then((files) => {
        console.log("received:" + files);
    }, (error) => {
        console.log(error);
    });
   
    
    console.log("Code submitted");

  }
  render() {
    const code = ''
    console.log('Type: ' + typeof window.submission)
    return (
    <div className="renderCode">
        <h2 style={{textAlign: 'center',}} >Circular Queue</h2>
        <h4 style={{textAlign: 'center',}}>Author:  </h4>
        <Button color='primary' className='px-4' onClick={this.redirectToComment}>
            Go To Comments
        </Button>
        <pre className="line-numbers">
            <code className="language-java">
            {window.submission.content}
            </code>
        </pre>
    </div>
    )
  }
}

export default Code;