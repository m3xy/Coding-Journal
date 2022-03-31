/**
 * Profile.jsx
 * author: 190019931
 *
 * A User's profile page
 */

import React, { useState, useEffect } from "react"
import { Tabs, Tab, ListGroup, Badge, Dropdown } from "react-bootstrap"
import { useNavigate, useParams } from "react-router-dom"
import axiosInstance from "../Web/axiosInstance"
import JwtService from "../Web/jwt.service"

const userEndpoint = "/user"
const userTypeEndpoint = "/changepermissions"
const submissionEndpoint = "/submission"

//User Types
const userTypes = [
	"User",
	"Publisher",
	"Reviewer",
	"Reviewer-Publisher", //Deprecated
	"Editor"
]

function Profile() {
	//Hook returns a navigate function used to navigate between
	const navigate = useNavigate()

	//If viewing another profile, ID of the user is fetched from the URL parameters
	const { id } = useParams()

	//The user the profile describes
	const [user, setUser] = useState({})

	//Returns the user's ID (to get their profile)
	const getUserID = () => {
		let user = id ? id : JwtService.getUserID()
		return user
	}

	//If no user is not logged in, navigate back to the login page
	if (getUserID() === null) {
		navigate("/login")
	}

	//Fetch the User's details from their ID. useEffect hook is invoked when the page (re)renders/dependency changes (User ID in this case)
	useEffect(() => {
		axiosInstance
			.get(userEndpoint + "/" + getUserID())
			.then((response) => {
				setUser(response.data)
			})
			.catch(() => {
				return <div></div>
			})
	}, [id])

	//Returns a list of submissions (used in displaying authored and reviewed submissions)
	const getSubmissionsList = (submissions) => {
		return (
			<ListGroup>
				{submissions?.map((submission) => {
					return (
						<ListGroup.Item
							as="li"
							key={submission.ID}
							style={{ padding: "10px" }}
							action
							onClick={() => {
								;(!id ||
									userTypes(JwtService.getUserType()) ==
										"Editor" ||
									submission.approved) &&
									navigate(
										submissionEndpoint + "/" + submission.ID
									)
							}}>
							<Badge
								bg={
									submission.approved
										? "primary"
										: submission.approved === null
										? "secondary"
										: "danger"
								}
								style={{ marginRight: "15px" }}
								pill>
								{submission.approved
									? "Approved"
									: submission.approved === null
									? "In review"
									: "Rejected"}
							</Badge>
							<label>{cutShort(submission.name, 50)}</label>
						</ListGroup.Item>
					)
				})}
			</ListGroup>
		)
	}

	// Shorten a string if it is too long for given format.
	const cutShort = (text, limit) => {
		if (text.length > limit) {
			let short = text.substring(0, limit)
			return (
				(short.includes(" ")
					? short.substring(0, short.lastIndexOf(" "))
					: short.substring(0, limit - 12)) + "..."
			)
		} else {
			return text
		}
	}

	//Sends a request to the backend to edit the user type of the user (Only editors can make this request)
	const editType = (type) => {
		axiosInstance
			.post(userEndpoint + "/" + getUserID() + userTypeEndpoint, {
				permissions: type
			})
			.then((response) => {
				setUser({ ...user, userType: type })
			})
			.catch((error) => {
				console.log(error)
			})
	}

	//A button which allows an Editor to modify another User's type
	const editTypeBtn = () => {
		return (
			<Dropdown>
				<Dropdown.Toggle variant="light" bsPrefix="p-0">
					✎
				</Dropdown.Toggle>
				<Dropdown.Menu>
					{userTypes.map((type, i) => {
						if (i == 3) return
						return (
							<Dropdown.Item onClick={() => editType(i)}>
								{type}
							</Dropdown.Item>
						)
					})}
				</Dropdown.Menu>
			</Dropdown>
		)
	}

	return (
		<div className="col-md-6 offset-md-3" style={{ textAlign: "center" }}>
			<br />
			<h2>{user.firstName + " " + user.lastName}</h2>
			<label>({userTypes[user.userType]})</label>
			{userTypes[JwtService.getUserType()] == "Editor" &&
				JwtService.getUserID() !== getUserID() &&
				editTypeBtn()}
			<br />
			<br />
			<Tabs
				justify
				defaultActiveKey="authored"
				id="profileTabs"
				className="mb-3">
				<Tab eventKey="authored" title="Authored Submissions">
					{user.authoredSubmissions?.length > 0 ? (
						getSubmissionsList(user.authoredSubmissions)
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No authored submissions</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="reviewed" title="Reviewed Submissions">
					{user.reviewedSubmissions?.length > 0 ? (
						getSubmissionsList(user.reviewedSubmissions)
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No reviewed submissions</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="contact" title="Contact">
					Email: {user.profile?.email} <br />
					Phone Number: {user.profile?.phoneNumber} <br />
					Organization: {user.profile?.organization} <br />
				</Tab>
			</Tabs>
		</div>
	)
}

export default Profile
