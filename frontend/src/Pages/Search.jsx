/**
 * Search.jsx
 * author: 190019931
 *
 * Search page
 */

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
	//search parameters passed to the page (most notably, the "name" search parameter is used in both user and submission search)
	const [search, setSearch] = useSearchParams()

	//Hook returns a navigate function used to navigate between
	const navigate = useNavigate()

	//Result submissions from submissions query
	const [submissions, setSubmissions] = useState([])

	//Result users from users query
	const [users, setUsers] = useState([])

	//The display order of submissions
	const [submissionOrder, setSubmissionOrder] = useState(null)

	//User type for users query
	const [userType, setUserType] = useState(null)

	//The display order of users from users query
	const [userOrder, setUserOrder] = useState(null)

	//Tags for submissions query
	const [tags, setTags] = useState([])

	//The tags input
	const [tagInput, setTagInput] = useState("")

	//The authors for submissions query
	const [authors, setAuthors] = useState([])

	//The reviewers for submissions query
	const [reviewers, setReviewers] = useState([])

	//Perform a submissions and users search - useEffect hook is invoked when the page (re)renders/dependency changes (query/options/filters)
	useEffect(() => {
		searchSubmissions()
		searchUsers()
	}, [search, submissionOrder, userOrder, tags, userType, authors, reviewers])

	//Search for submissions with the specified options (query/options/filters)
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

	//Get each submissions details by their submission ID returned from the submissions query
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

	//Search users with the specified options (query/options/filters)
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

	//Return all of the submission cards of the submissions query (Maps each submission to a SubmissionCard component to display)
	const submissionCards = submissions?.map((submission, i) => {
		return <SubmissionCard submission={submission} key={i} />
	})

	//Adds a user to the submissions search filter (as an author/reviewer)
	const addUser = (user, users, setUsers, type) => {
		if (!users.some((elem) => elem.userId == user.userId))
			return (
				<Card.Link
					size="sm"
					onClick={() => setUsers((users) => [...users, user])}>
					Add as {type}
				</Card.Link>
			)
	}

	//Return all of the user cards of the users query (Maps each user to a Card component to display)
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

	//(Removable) Buttons displaying the tags of the submissions query
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

	//Container for (user and submission) cards (results of the users and submissions queries)
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

	//Authors and Reviewers can be added as filters to the submissions query, these are displayed as removable buttons, alike tags
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
				<option value={4}>Editor</option>
			</Form.Select>
			<br />
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
				<div>
					Authors: {userBtns(authors, setAuthors)} <br />
				</div>
			)}
			{reviewers.length > 0 && (
				<div>
					Reviewers: {userBtns(reviewers, setReviewers)} <br />
				</div>
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
