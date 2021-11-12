/**
 * Profile.jsx
 * author: 190019931
 * 
 * User's profile page
 */
import React from "react";
import {Tabs, Tab, ListGroup} from "react-bootstrap";
import { Redirect } from "react-router-dom";

class Profile extends React.Component {

	constructor(props) {
        super(props);

        this.state = {
			userID: this.getUserID(),
			firstname: "",
			lastname: "",
			email: "",
			usertype: 0,
			phonenumber: "",
			organization: "",
			projects: {}
        };

		this.logout= this.logout.bind(this);
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

	openProject(projectID) {
		var codePage = window.open("/code");
		codePage.projectID = projectID;
	}

	componentDidMount() {
		this.props.getProfile(this.state.userID).then((user) => {
			console.log(user.projects);
			this.setState({
				firstname: user.firstname,
				lastname: user.lastname,
				usertype: user.usertype,
				email: user.email,
				phonenumber: user.phonenumber,
				organization: user.organization,
				projects: user.projects
			});
		})
	}

	render() {

		//Get user details from ID
		// const firstName = "John";
		// const lastName = "Doe";
		// const userType = "User";
		// const email = "JohnDoe@gmail.com"

		//Get user posts
		// const posts = []

		//Get user comments
		const comments = []

		const userTypes = ["None", "Publisher", "Reviewer", "Reviewer-Publisher", "User"]

		const projects = Object.entries(this.state.projects).map(([id, name]) => {
			return (
                <ListGroup.Item as="li" key={id} action onClick={() => {this.openProject(id)}}>
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
						{projects.length > 0 ? (
							<ListGroup>{projects}</ListGroup>
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