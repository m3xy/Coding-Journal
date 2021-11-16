/**
 * Home.jsx
 * author: 190010714
 * 
 * Home page for the whole website
 */

import React from 'react';
import { Button } from 'react-bootstrap';
import  { Redirect } from 'react-router-dom'

class Home extends React.Component {

    constructor(props){
        super(props);

        this.state = {
            userID: this.getUserID()
        }

        this.logout= this.logout.bind(this);
    }

    componentDidMount() {
        
    }

    getUserID() {
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userID"){       //  If userID key exists, extract the userID value
                return JSON.parse(cookie[1].trim()).userId;
            }
        }
        return null;
    }

    logout() {
        var cookies = document.cookie.split(';'); 
    
        // The "expire" attribute of every cookie is set to "Thu, 01 Jan 1970 00:00:00 GMT".
        for (var i = 0; i < cookies.length; i++) {
            document.cookie = cookies[i] + "=;expires=" + new Date(0).toUTCString();  //Setting all cookies expiry date to be a past date.
        }

        this.setState({
            userID : null
        })
    }

    render() {
    
        console.log(this.state.userID);

        if(this.getUserID() === null) {
            return (<Redirect to ='/login' />);
        }

        return (
            <div className="text-center">
                <br/>
                You are logged in.
                <br/>
                <Button variant="outline-danger" onClick={this.logout}>Logout</Button>{' '}
            </div>
        );
    }
}

export default Home;