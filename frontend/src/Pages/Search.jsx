import React, { useState, useEffect } from "react"
import {
	Button,
	FormControl,
	InputGroup,
	Container,
	Row,
	Card,
	CardGroup,
	Form
} from "react-bootstrap"
import { useNavigate, useSearchParams } from "react-router-dom"
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
	const [submissionOrder, setSubmissionOrder] = useState(null)
	const [userType, setUserType] = useState(null)
	const [userOrder, setUserOrder] = useState(null)
	const [tags, setTags] = useState([])
	const [tagInput, setTagInput] = useState([])
	const [authors, setAuthors] = useState([])
	const [reviewers, setReviewers] = useState([])

	useEffect(() => {
		searchSubmissions()
		searchUsers()
	}, [search, submissionOrder, userOrder, tags, userType, authors, reviewers])

	const searchSubmissions = () => {
		setSubmissions([])

		const submissionSearch = new URLSearchParams(search)
		submissionOrder && submissionSearch.append("orderBy", submissionOrder)
		tags.forEach((tag) => {
			submissionSearch.append("tags", tag)
		})
		authors.forEach((author) => {
			submissionSearch.append("authors", author.userId)
		})
		reviewers.forEach((reviewer) => {
			submissionSearch.append("reviewers", reviewer.userId)
		})
		axiosInstance
			.get(submissionsQueryEndpoint, {
				params: submissionSearch
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
		const userSearch = new URLSearchParams(search)
		userType && userSearch.append("userType", userType)
		userOrder && userSearch.append("orderBy", userOrder)
		axiosInstance
			.get(usersQueryEndpoint, {
				params: userSearch
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

	const addUser = (user, users, setUsers, type) => {
		if(!users.some((elem) => elem.userId == user.userId))
			return (
				<Card.Link
					size="sm"
					onClick={() => setUsers((users) => [...users, user])}>
					Add as {type}
				</Card.Link>
			)
	}

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
					<Card.Link
						size="sm"
						onClick={() =>
							navigate(profileURL + "/" + user.userId)
						}>
						View
					</Card.Link>
					{addUser(user, authors, setAuthors, "author")}
					{addUser(user, reviewers, setReviewers, "reviewer")}
				</Card.Body>
				<Card.Footer>
					<small className="text-muted">
						Registered: {new Date(user.createdAt).toDateString()}
					</small>
				</Card.Footer>
			</Card>
		)
	})

	const tagBtns = tags.map((tag, i) => {
		return (
			<Button
				key={i}
				variant="outline-danger"
				size="sm"
				onClick={() => setTags(tags.filter((value) => value !== tag))}>
				{tag}
			</Button>
		)
	})

	const cardContainer = (cards) => {
		return (
			<CardGroup>
				<div
					style={{
						display: "flex",
						flexDirection: "row",
						overflowX: "auto",
						overflowY: "hidden"
					}}>
					{cards}
				</div>
			</CardGroup>
		)
	}

	const userBtns = (users, setUsers) =>
		users.map((user) => {
			return (
				<Button
					key={user.userId}
					variant="outline-danger"
					size="sm"
					onClick={() =>
						setUsers(
							users.filter((elem) => elem.userId !== user.userId)
						)
					}>
					{user.firstName +
						" " +
						user.lastName +
						" (" +
						user.profile.email +
						")"}
				</Button>
			)
		})

	return (
		<Container>
			<h3>Users:</h3>
			Sort By:
			<Form.Select onChange={(e) => setUserOrder(e.target.value)}>
				<option default value={null}></option>
				<option value="firstName">First Name</option>
				<option value="lastName">Last Name</option>
			</Form.Select>
			User Type:
			<Form.Select onChange={(e) => setUserType(e.target.value)}>
				<option default value={null}></option>
				<option value={0}>User</option>
				<option value={1}>Publisher</option>
				<option value={2}>Reviewer</option>
				<option value={3}>Reviewer-Publisher</option>
				<option value={4}>Editor</option>
			</Form.Select>
			<Row>
				{users && users.length > 0 ? (
					cardContainer(userCards)
				) : (
					<div>No Users found</div>
				)}
			</Row>
			<br />
			<h3>Submissions:</h3>
			Sort By:
			<Form.Select onChange={(e) => setSubmissionOrder(e.target.value)}>
				<option default value={null}></option>
				<option value="oldest">Newest</option>
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
						!tags.includes(tagInput) &&
							setTags((tags) => [...tags, tagInput])
						setTagInput("")
					}}>
					Add
				</Button>
			</InputGroup>
			{tagBtns}
			<br />
			{authors.length > 0 && (
				<>
					Authors: {userBtns(authors, setAuthors)} <br />
				</>
			)}
			{reviewers.length > 0 && (
				<>
					Reviewers: {userBtns(reviewers, setReviewers)} <br />
				</>
			)}
			<Row>
				{submissions && submissions.length > 0 ? (
					cardContainer(submissionCards)
				) : (
					<div>No submissions found</div>
				)}
			</Row>
		</Container>
	)
}

export default Search
