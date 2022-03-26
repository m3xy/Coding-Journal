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
	const [user, setUser] = useState({})

	if (getUserID() === null) {
		navigate("/login")
	}

	useEffect(() => {
		axiosInstance
			.get(profileEndpoint + "/" + getUserID())
			.then((response) => {
				setUser(response.data)
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
				{user.authoredSubmissions?.map((submission) => {
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
			<h2>{user.firstName + " " + user.lastName}</h2>
			<label>({userTypes[user.userType]})</label>
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
				<Tab eventKey="reviewed" title="Reviewer Submissions">
					{user.reviewedSubmissions?.length > 0 ? (
						getSubmissionsList(user.reviewedSubmissions)
					) : (
						<div className="text-center" style={{ color: "grey" }}>
							<i>No reviewed submissions</i>
						</div>
					)}
				</Tab>
				<Tab eventKey="comments" title="Comments">
					{}
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
