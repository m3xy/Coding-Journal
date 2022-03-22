import React, { useState, useEffect } from "react"
import { Card, Button, Badge } from "react-bootstrap"
import styles from "./SubmissionList.module.css"
import { useNavigate } from "react-router-dom"
import axiosInstance from "../../Web/axiosInstance"

export default ({ query, display }) => {
	const [submissions, setSubmissions] = useState([])
	const [authors, setAuthors] = useState({})
	const [error, setError] = useState(null)
	const navigate = useNavigate()

	useEffect(() => {
		axiosInstance
			.get("/submissions/query", {
				params: query
			})
			.then((response) => {
				if (response.data.hasOwnProperty(submissions))
					setSubmissionsFromPrimitives(response.data.submissions)
			})
			.catch((err) => {
				console.log(err)
				setError(true)
			})
	}, [])

	// Get submission details from their basic types.
	const setSubmissionsFromPrimitives = (submissionPrimitives) => {
		submissionPrimitives.map((submission) => {
			axiosInstance
				.get("/submission/" + submission.ID)
				.then((response) => {
					setSubmissions((submissions) => {
						return [...submissions, response.data]
					})
					setAuthorsNotFetched(response.data.authors)
				})
				.catch((err) => {
					console.log(err)
				})
		})
	}

	// Get authors from their IDs.
	const setAuthorsNotFetched = (authors) => {
		localAuthors = {}
		authors.map((author) => {
			if (!localAuthors.hasOwnProperty(author.userId)) {
				axiosInstance
					.get("/user/" + author.userId)
					.then((response) => {
						localAuthors = {
							...localAuthors,
							[author.userId]: response.data
						}
					})
					.catch((err) => {
						console.log(err)
					})
			}
		})
		setAuthors(localAuthors)
	}

	const getBadge = (submission) => {
		return (
			<Badge
				bg={
					submission.approved
						? "primary"
						: submission.approved === null
						? "secondary"
						: "danger"
				}
				pill>
				{submission.approved
					? "Approved"
					: submission.approved === null
					? "In review"
					: "Rejected"}
			</Badge>
		)
	}

	const getAuthorFullName = (author) => {
		if (authors.hasOwnProperty(author.userId)) {
			return (
				authors[author.userId].profile.firstName +
				" " +
				authors[author.userId].profile.lastName
			)
		} else {
			return author.userId
		}
	}

	const getTags = (submission) => {
		if (submission.hasOwnProperty("categories"))
			return submission.categories.map((category, i) => {
				return (
					<Button
						key={i}
						variant="outline-secondary"
						size="sm"
						disabled
						style={{ margin: "3px" }}>
						{category.category}
					</Button>
				)
			})
	}

	// Get a list of cards from the list of submissions.
	const getCards = (submissions) => {
		return submissions.map((submission, i) => {
			return (
				<Card
					key={i}
					style={{
						minWidth: "25rem",
						maxWidth: "25rem",
						margin: "8px"
					}}
					className="shadow rounded">
					<Card.Body>
						<Card.Subtitle className="mb-2">
							{getBadge(submission)}
						</Card.Subtitle>
						<Card.Title>{cutShort(submission.name, 40)}</Card.Title>
						<Card.Subtitle className="mb-2 text-muted">
							{" "}
							{submission.authors.length > 1
								? "Authors:"
								: "Author:"}{" "}
							{submission.authors.map((author, index) => {
								return (
									(index === 0 ? " " : ", ") +
									getAuthorFullName(author)
								)
							})}{" "}
						</Card.Subtitle>
					</Card.Body>
					<Card.Body
						style={{
							height: "60%",
							whiteSpace: "normal"
						}}>
						<Card.Text>
							{cutShort(submission.metaData.abstract, 200)}
						</Card.Text>
					</Card.Body>
					<Card.Subtitle>{getTags(submission)}</Card.Subtitle>
					<Card.Body>
						<Button
							variant="primary"
							onClick={() => {
								navigate("/submission/" + submission.ID)
							}}>
							Explore
						</Button>
					</Card.Body>
					<Card.Footer className="text-muted">
						Created: {new Date(submission.CreatedAt).toDateString()}
					</Card.Footer>
				</Card>
			)
		})
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
		<div className={styles.scrollerComponent}>
			<h2>{display}</h2>
			<div className={styles.ScrollerContainer}>
				{error
					? "Something went wrong, please try again later..."
					: submissions.length > 0
					? getCards(submissions)
					: "It's empty here..."}
			</div>
		</div>
	)
}
