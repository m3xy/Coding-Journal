/**
 * Home.jsx
 * author: 190010714
 * 
 * Home page for the whole website
 */

import React from 'react';
import  { Redirect } from 'react-router-dom'

class Home extends React.Component {

    constructor(props){
        super(props);

        this.state = {
            userID: null
        }
    }

    componentDidMount() {
        
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userID"){       //  If userID key exists, extract the userID value
                this.setState({
                    userID: cookie[1].trim()
                })
                break;
            }
        }
    }

    render() {

        const userLoggedIn = this.state.userID;

        if(userLoggedIn === null) {
            return (<Redirect to ='/login' />);
        }

        return (
            <div className="col-md-6 col-md-offset-3">
                Logged in.
            </div>
        );
    }
}

export default Home;