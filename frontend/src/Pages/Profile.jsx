/**
 * Profile.jsx
 * author: 190019931
 *
 * User's profile page
 */

import React, { useState, useEffect } from "react"
import { Tabs, Tab, ListGroup, Badge } from "react-bootstrap"
import { useNavigate } from "react-router-dom"
import axiosInstance from "../Web/axiosInstance"
import JwtService from "../Web/jwt.service"

const profileEndpoint = "/user"

function getUserID() {
	let user = JwtService.getUserID()
	return user
}

function Profile() {
	const navigate = useNavigate()
	const [firstname, setFirstname] = useState("")
	const [lastname, setLastname] = useState("")
	const [usertype, setUsertype] = useState(0)
	const [email, setEmail] = useState("")
	const [phonenumber, setPhoneNumber] = useState("000000")
	const [organization, setOrganization] = useState("None")
	const [authoredSubs, setAuthoredSubs] = useState([])
	const [reviewedSubs, setReviewedSubs] = useState([])

	if (getUserID() === null) {
		navigate("/login")
	}

	useEffect(() => {
		axiosInstance
			.get(profileEndpoint + "/" + getUserID())
			.then((response) => {
				console.log(response.data)
				setFirstname(response.data.profile.firstName)
				setLastname(response.data.profile.lastName)
				setUsertype(response.data.userType)
				setEmail(response.data.profile.email)
				setPhoneNumber(response.data.profile.phoneNumber)
				setOrganization(response.data.profile.organization)
				setAuthoredSubs(
					response.data.authoredSubmissions
						? response.data.authoredSubmissions
						: []
				)
				setReviewedSubs(
					response.data.reviewedSubmissions
						? response.data.reviewedSubmissions
						: []
				)
			})
			.catch(() => {
				return <div></div>
			})
	}, [])

	//Get user comments
	const comments = []
	const userTypes = [
		"User",
		"Publisher",
		"Reviewer",
		"Reviewer-Publisher",
		"Editor"
	]

	const getSubmissionsList = (submissions) => {
		return (
			<ListGroup>
				{submissions.map((submission) => {
					return (
						<ListGroup.Item
							as="li"
							key={submission.ID}
							style={{ padding: "10px" }}
							action
							onClick={() => {
								navigate("/submission/" + submission.ID)
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

	return (
		<div className="col-md-6 offset-md-3" style={{ textAlign: "center" }}>
			<br />
			<h2>{firstname + " " + lastname}</h2>
			<label>({userTypes[usertype]})</label>
			<br />
			<br />
			<Tabs
				justify
				defaultActiveKey="authored"
				id="profileTabs"
				className="mb-3">
				<Tab eventKey="authored" title="Authored Submissions">
					{authoredSubs.length > 0 ? (
						getSubmissionsList(authoredSubs)
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No authored submissions</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="reviewed" title="Reviewer Submissions">
					{reviewedSubs.length > 0 ? (
						getSubmissionsList(reviewedSubs)
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No reviewed submissions</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="comments" title="Comments">
					{comments}
				</Tab>
				<Tab eventKey="contact" title="Contact">
					Email: {email} <br />
					Phone Number: {phonenumber} <br />
					Organization: {organization} <br />
				</Tab>
			</Tabs>
		</div>
	)
}

export default Profile
