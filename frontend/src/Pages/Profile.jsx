/**
 * Profile.jsx
 * author: 190019931
 *
 * A User's profile page
 */

import React, { useState, useEffect } from "react"
import { Tabs, Tab, ListGroup, Badge, Card, Col, Row } from "react-bootstrap"
import { useNavigate, useParams } from "react-router-dom"

import axiosInstance from "../Web/axiosInstance"
import JwtService from "../Web/jwt.service"
import AnalyticsAndSettings from "../Components/AnalyticsAndSettings"

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
	const [user, setUser] = useState({
		userId: "",
		firstName: "",
		lastName: "",
		profile: {},
		authoredSubmissions: []
	})
	let [numAccepted, setNumAccepted] = useState(0)
	let [numUnderReview, setNumUnderReview] = useState(0)
	let [numRejected, setNumRejected] = useState(0)

	//Returns the user's ID (to get their profile)
	const getUserID = () => {
		let user = id ? id : JwtService.getUserID()
		return user
	}

	//If no user is not logged in, navigate back to the login page
	if (getUserID() === null) {
		navigate("/login")
	}

	const updateUser = () => {
		axiosInstance
			.get(userEndpoint + "/" + getUserID())
			.then((response) => {
				setUser(response.data)
			})
			.catch(() => {
				return <div></div>
			})
	}

	useEffect(() => {
		updateUser()
	}, [])

	useEffect(() => {
		user.authoredSubmissions?.map((submission) => {
			submission?.approved
				? setNumAccepted((numAccepted) => ++numAccepted)
				: submission?.approved === null
				? setNumUnderReview((numUnderReview) => ++numUnderReview)
				: setNumRejected((numRejected) => ++numRejected)
		})
	}, [user])

	//Get user comments
	const comments = []
	const userTypes = [
		"User",
		"Publisher",
		"Reviewer",
		"Reviewer-Publisher",
		"Editor"
	]

	const getBadge = (approved) => {
		const [bg, text] = approved
			? ["primary", "Approved"]
			: approved === null
			? ["secondary", "In Review"]
			: ["danger", "Rejected"]

		return <Badge bg={bg}>{text}</Badge>
	}

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
									userTypes[JwtService.getUserType()] ==
										"Editor" ||
									submission.approved) &&
									navigate(
										submissionEndpoint + "/" + submission.ID
									)
							}}>
							{getBadge(submission.approved)}
							<label>{cutShort(submission.name, 50)}</label>
						</ListGroup.Item>
					)
				})}
			</ListGroup>
		)
	}

	const getAnalytics = () => {
		return (
			<Card>
				<Col>
					<Card
						border="primary"
						style={{ width: "flex", margin: "8px" }}>
						<Card.Header>Total Submissions</Card.Header>
						<Card.Body>
							<Card.Title>
								{numAccepted + numUnderReview + numRejected}
							</Card.Title>
						</Card.Body>
					</Card>
				</Col>
				<Col>
					<Card
						border="success"
						style={{ width: "flex", margin: "8px" }}>
						<Card.Header>Accepted Submissions</Card.Header>
						<Card.Body>
							<Card.Title>{numAccepted}</Card.Title>
						</Card.Body>
					</Card>
				</Col>
				<Col>
					<Card
						border="warning"
						style={{ width: "flex", margin: "8px" }}>
						<Card.Header>Submissions Under-Review</Card.Header>
						<Card.Body>
							<Card.Title>{numUnderReview}</Card.Title>
						</Card.Body>
					</Card>
				</Col>
				<Col>
					<Card
						border="danger"
						style={{ width: "flex", margin: "8px" }}>
						<Card.Header>Rejected Submissions</Card.Header>
						<Card.Body>
							<Card.Title>{numRejected}</Card.Title>
						</Card.Body>
					</Card>
				</Col>
			</Card>
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
							<i>No reviewed Submission</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="analytics" title="Analytics">
					{user.authoredSubmissions?.length > 0 ? (
						getAnalytics()
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No Data to Analyze</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="contact" title="Contact">
					Email: {user.profile?.email} <br />
					Phone Number: {user.profile?.phoneNumber} <br />
					Organization: {user.profile?.organization} <br />
				</Tab>
				<Tab eventKey="edit" title="Edit Profile">
					<AnalyticsAndSettings updateUser={updateUser} />
				</Tab>
			</Tabs>
		</div>
	)
}

export default Profile
