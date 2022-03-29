import React, { useState, useEffect } from "react"
import {
	Button,
	Collapse,
	FormControl,
	InputGroup,
	Toast,
	Container,
	Row,
	Col,
	Card,
	CardGroup
} from "react-bootstrap"
import {
	useNavigate,
	useSearchParams,
	createSearchParams
} from "react-router-dom"
import axiosInstance from "../Web/axiosInstance"
import SubmissionCard from "../Components/SubmissionCard"

const submissionsQueryEndpoint = "/submissions/query"
const usersQueryEndpoint = "/users/query"
const submissionEndpoint = "/submission"
const profileURL = "/profile"

const userTypes = [
	"User",
	"Publisher",
	"Reviewer",
	"Reviewer-Publisher",
	"Editor"
]

function Search() {
	const [search, setSearch] = useSearchParams()
	const navigate = useNavigate()
	const [submissions, setSubmissions] = useState([])
	const [users, setUsers] = useState([])
	const [showSubmissions, setShowSubmissions] = useState(false)
	const [showUsers, setShowUsers] = useState(false)

	useEffect(() => {
		setSubmissions([])
		setUsers([])
		searchSubmissions()
		searchUsers()
	}, [search])

	const searchSubmissions = () => {
		axiosInstance
			.get(submissionsQueryEndpoint, { params: search })
			.then((response) => {
				if (response.data.hasOwnProperty("submissions"))
					getSubmissions(response.data.submissions)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const getSubmissions = (response) =>
		response?.map((submission) => {
			axiosInstance
				.get(submissionEndpoint + "/" + submission.ID)
				.then((response) => {
					setSubmissions((submissions) => {
						return [...submissions, response.data]
					})
				})
				.catch((error) => {
					console.log(error)
				})
		})

	const searchUsers = () => {
		axiosInstance
			.get(usersQueryEndpoint, { params: search })
			.then((response) => {
				setUsers(response.data.users)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const submissionCards = submissions?.map((submission, i) => {
		return <SubmissionCard submission={submission} key={i} />
	})

	const userCards = users?.map((user, i) => {
		return (
			<Card
				key={i}
				style={{ minWidth: "45%", maxWidth: "45%", margin: "8px" }}>
				<Card.Header closeButton={false}>
					{userTypes[user.userType]}
				</Card.Header>
				<Card.Body>
					<Card.Title>
						{user.firstName + " " + user.lastName}
					</Card.Title>
					<Card.Subtitle className="mb-2 text-muted">
						{user.profile.email}
					</Card.Subtitle>
					<br />
					<Button
						variant="dark"
						size="sm"
						onClick={() =>
							navigate(profileURL + "/" + user.userId)
						}>
						View
					</Button>
				</Card.Body>
				<Card.Footer>
					<small className="text-muted">
						Registered: {new Date(user.createdAt).toDateString()}
					</small>
				</Card.Footer>
			</Card>
		)
	})

	return (
		<Row>
			<Col className="span-12">
				{users && users.length > 0 ? (
					<CardGroup>{userCards}</CardGroup>
				) : (
					<div>No Users found</div>
				)}
			</Col>
			<Col className="span-12">
				{submissions && submissions.length > 0 ? (
					<CardGroup>{submissionCards}</CardGroup>
				) : (
					<div>No submissions found</div>
				)}
			</Col>
		</Row>
	)
}

export default Search
