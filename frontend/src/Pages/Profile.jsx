/**
 * Profile.jsx
 * author: 190019931
 * 
 * User's profile page
 */
import React from "react";
import {Tabs, Tab, ListGroup} from "react-bootstrap";
import { Redirect } from "react-router-dom";
import axiosInstance from "../Web/axiosInstance";

const profileEndpoint = '/users';

class Profile extends React.Component {

	constructor(props) {
        super(props);

        this.state = {
			userId: this.getUserID(),
			firstname: "",
			lastname: "",
			email: "",
			usertype: 0,
			phonenumber: "",
			organization: "",
			submissions: {}
        };

		this.logout= this.logout.bind(this);
    }

	getUserID() {
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userId") {       //  If userId key exists, extract the userId value
                return cookie[1].trim();
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
            userId : null
        })
    }

	openSubmission(submissionID) {
		var codePage = window.open("/code");
		codePage.submissionID = submissionID;
	}

	componentDidMount() {
        axiosInstance.get(profileEndpoint + "/" + this.state.userId)
		             .then((response) => {
			console.log(response.data);
			this.setState({
				firstname: response.data.firstname,
				lastname: response.data.lastname,
				usertype: response.data.usertype,
				email: response.data.email,
				phonenumber: response.data.phonenumber,
				organization: response.data.organization,
				submissions: response.data.submissions
			});
		})
	}

	render() {
		//Get user comments
		const comments = []

		const userTypes = ["None", "Publisher", "Reviewer", "Reviewer-Publisher", "User"]

		const submissions = Object.entries(this.state.submissions).map(([id, name]) => {
			return (
                <ListGroup.Item as="li" key={id} action onClick={() => {this.openSubmission(id)}}>
                    <label>{name}</label>
                </ListGroup.Item>
            );
		})

		if(this.getUserID() === null) {
            return (<Redirect to ='/login' />);
        }

		return (
			<div className="col-md-6 offset-md-3" style={{textAlign:"center"}}>
				<br/>
				<h2>{this.state.firstname + " " + this.state.lastname}</h2>
				 <label>({userTypes[this.state.usertype]})</label>
				<br/><br/>
				<Tabs justify defaultActiveKey="profile" id="profileTabs" className="mb-3">
					<Tab eventKey="posts" title="Posts">
						{submissions.length > 0 ? (
							<ListGroup>{submissions}</ListGroup>
						) : (
							<div className="text-center" style={{color:"grey"}}><i>No posts</i></div>
						)	
						}
					</Tab>
					<Tab eventKey="comments" title="Comments">
						{comments}
					</Tab>
					<Tab eventKey="contact" title="Contact">
						Email: {this.state.email} <br/>
						Phone Number: {this.state.phonenumber}  <br/>
						Organization: {this.state.organization} <br/>
					</Tab>
				</Tabs>
			</div>
		);

	}
}

export default Profile;
