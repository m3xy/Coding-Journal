/**
 * Profile.jsx
 * author: 190019931
 * 
 * User's profile page
 */
import React, { useState, useEffect } from "react";
import {Tabs, Tab, ListGroup} from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import axiosInstance from "../Web/axiosInstance";

const profileEndpoint = '/users';

function getUserID() {
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userId") {       //  If userId key exists, extract the userId value
                return cookie[1].trim();
            }
        }
        return null;
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


    function openSubmission(submissionsID) {
        navigate("/code/" + submissionsID)
    }

	if(getUserID() === null) {
		navigate("/login")
	}

    useEffect(() => {
		axiosInstance.get(profileEndpoint + "/" + getUserID())
					.then((response) => {
						console.log(response.data);
						setFirstname(response.data.firstname)
						setLastname(response.data.lastname)
						setUsertype(response.data.usertype)
						setEmail(response.data.email)
						setPhoneNumber(response.data.phonenumber)
						setOrganization(response.data.organization)
						setSubmissions(response.data.submissions)
					})
					.catch(() => {
						return(<div></div>)
					})
    }, [])

	//Get user comments
	const comments = []
	const userTypes = ["None", "Publisher", "Reviewer", "Reviewer-Publisher", "User"]
	const submissions = Object.entries(userSubmissions).map(([id, name]) => {
		return (
			<ListGroup.Item as="li" key={id} action onClick={() => {openSubmission(id)}}>
				<label>{name}</label>
			</ListGroup.Item>
		);
	})
	return (
		<div className="col-md-6 offset-md-3" style={{textAlign:"center"}}>
			<br/>
			<h2>{firstname + " " + lastname}</h2>
			<label>({userTypes[usertype]})</label>
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
					Email: {email} <br/>
					Phone Number: {phonenumber}  <br/>
					Organization: {organization} <br/>
				</Tab>
			</Tabs>
		</div>
	)
}

export default Profile;
