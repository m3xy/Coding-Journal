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
	CardGroup,
	Form
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

const defaultSubmissionOrder = "oldest"
const defaultUserOrder = "firstName"

function Search() {
	const [search, setSearch] = useSearchParams()
	const navigate = useNavigate()
	const [submissions, setSubmissions] = useState([])
	const [users, setUsers] = useState([])
	const [showSubmissions, setShowSubmissions] = useState(false)
	const [showUsers, setShowUsers] = useState(false)
	const [submissionOrder, setSubmissionOrder] = useState(
		defaultSubmissionOrder
	)
	const [userOrder, setUserOrder] = useState(defaultUserOrder)
	const [tags, setTags] = useState([])
	const [tagInput, setTagInput] = useState([])

	useEffect(() => {
		searchSubmissions()
		searchUsers()
	}, [search, submissionOrder, userOrder, tags])

	const searchSubmissions = () => {
		setSubmissions([])
		axiosInstance
			.get(submissionsQueryEndpoint, {
				params: { name: search.get("name"), tags: tags, orderBy: submissionOrder }
			})
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
		setUsers([])
		axiosInstance
			.get(usersQueryEndpoint, {
				params: { name: search.get("name"), orderBy: userOrder }
			})
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
				style={{ minWidth: "25rem", maxWidth: "25rem", margin: "8px" }}>
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

	const tagButtons = tags.map((tag, i) => {
		return (
			<Button
				key={i}
				variant="outline-secondary"
				size="sm"
				onClick={() => setTags(tags.filter((value) => value !== tag))}>
				{tag}
			</Button>
		)
	})

	return (
		<Container>
			<h3>Users:</h3>
			Sort By:
			<Form.Select onChange={(e) => setUserOrder(e.target.value)}>
				<option default value="firstName">
					First Name
				</option>
				<option value="lastName">Last Name</option>
			</Form.Select>
			<Row>
				{users && users.length > 0 ? (
					<CardGroup>
						{" "}
						<div
							style={{
								display: "flex",
								flexDirection: "row",
								overflowX: "auto",
								overflowY: "hidden"
							}}>
							{userCards}
						</div>
					</CardGroup>
				) : (
					<div>No Users found</div>
				)}
			</Row>
			<br />
			<h3>Submissions:</h3>
			Sort By:
			<Form.Select onChange={(e) => setSubmissionOrder(e.target.value)}>
				<option default value="oldest">
					Newest
				</option>
				<option value="newest">Oldest</option>
				<option value="alphabetical">Alphabetical</option>
			</Form.Select>
			Tags:
			<InputGroup className="mb-3">
				<FormControl
					placeholder="Add tags here"
					onChange={(e) => {
						setTagInput(e.target.value)
					}}
					value={tagInput}
				/>
				<Button
					variant="outline-secondary"
					onClick={() => {
						!tags.includes(tagInput) && setTags(tags => [...tags, tagInput])
						setTagInput("")
					}}>
					Add
				</Button>
			</InputGroup>
			{tagButtons}
			<Row>
				{submissions && submissions.length > 0 ? (
					<CardGroup>
						{" "}
						<div
							style={{
								display: "flex",
								flexDirection: "row",
								overflowX: "auto",
								overflowY: "hidden"
							}}>
							{submissionCards}
						</div>
					</CardGroup>
				) : (
					<div>No submissions found</div>
				)}
			</Row>
		</Container>
	)
}

export default Search
