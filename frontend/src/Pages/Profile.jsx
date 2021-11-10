/**
 * Profile.jsx
 * author: 190019931
 * 
 * User's profile page
 */
import React from "react";
import {Tabs, Tab} from "react-bootstrap";

class Profile extends React.Component {

	constructor(props) {
        super(props);

        this.state = {
			userId: 0,
            loggedIn: false
        };
    }

	componentDidMount() {
		
	}

	render() {
		//Get user details from ID
		const firstName = "John";
		const lastName = "Doe";
		const userType = "User";
		const email = "JohnDoe@gmail.com"

		//Get user posts
		const posts = []

		//Get user comments
		const comments = []

		return (
			<div className="col-md-6 offset-md-3" style={{textAlign:"center"}}>
				<br/>
				<h2>{firstName + " " + lastName}</h2>
			 	<label>({userType})</label>
				<br/><br/>
				<Tabs justify defaultActiveKey="profile" id="profileTabs" className="mb-3">
					<Tab eventKey="posts" title="Posts">
						{posts}
					</Tab>
					<Tab eventKey="comments" title="Comments">
						{comments}
					</Tab>
					<Tab eventKey="contact" title="Contact">
						Email: {email}
					</Tab>
				</Tabs>
			</div>
		);
	}
}

export default Profile;