/**
 * Profile.jsx
 * author: 190019931
 *
 * User's profile page
 */

import React, { useState, useEffect } from "react";
import { Tabs, Tab, ListGroup } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import axiosInstance from "../Web/axiosInstance";
import JwtService from "../Web/jwt.service";

const profileEndpoint = '/user';

function getUserID() {
	let user = JwtService.getUserID();
	return user;
}

function Profile() {
	const navigate = useNavigate()
	const [firstname, setFirstname] = useState('')
	const [lastname, setLastname] = useState('')
	const [usertype, setUsertype] = useState(0)
	const [email, setEmail] = useState('')
	const [phonenumber, setPhoneNumber] = useState('000000')
	const [organization, setOrganization] = useState('None')
	const [userSubmissions, setSubmissions] = useState('')


	function openSubmission(submissionsID, submissionName) {
		navigate("/code/" + submissionsID + "/" + submissionName)
	}

	if (getUserID() === null) {
		navigate("/login")
	}

	useEffect(() => {
		axiosInstance.get(profileEndpoint + "/" + getUserID())
			.then((response) => {
				console.log(response.data);
				setFirstname(response.data.profile.firstName)
				setLastname(response.data.profile.lastName)
				setUsertype(response.data.userType)
				setEmail(response.data.profile.email)
				setPhoneNumber(response.data.profile.phoneNumber)
				setOrganization(response.data.profile.organization)
				// setSubmissions(response.data.submissions)
			})
			.catch(() => {
				return (<div></div>)
			})
	}, [])

	//Get user comments
	const comments = []
	const userTypes = ["None", "Publisher", "Reviewer", "Reviewer-Publisher", "User"]
	const submissions = Object.entries(userSubmissions).map(([id, name]) => {
		return (
			<ListGroup.Item as="li" key={id} action onClick={() => { openSubmission(id, name) }}>
				<label>{name}</label>
			</ListGroup.Item>
		);
	})
	return (
		<div className="col-md-6 offset-md-3" style={{ textAlign: "center" }}>
			<br />
			<h2>{firstname + " " + lastname}</h2>
			<label>({userTypes[usertype]})</label>
			<br /><br />
			<Tabs justify defaultActiveKey="profile" id="profileTabs" className="mb-3">
				<Tab eventKey="posts" title="Posts">
					{submissions.length > 0 ? (
						<ListGroup>{submissions}</ListGroup>
					) : (
						<div className="text-center" style={{ color: "grey" }}><i>No posts</i></div>
					)
					}
				</Tab>
				<Tab eventKey="comments" title="Comments">
					{comments}
				</Tab>
				<Tab eventKey="contact" title="Contact">
					Email: {email} <br />
					Phone Number: {phonenumber}  <br />
					Organization: {organization} <br />
				</Tab>
			</Tabs>
		</div>
	)
}

export default Profile;
