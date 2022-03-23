import React, { useEffect, useState } from "react"
import axiosInstance from "../../Web/axiosInstance"
import { Card, Button, Badge } from "react-bootstrap"
import styles from "./SubmissionCard.module.css"
import { useNavigate } from "react-router-dom"

export default ({ submission }) => {
	const [authors, setAuthors] = useState([])
	const navigate = useNavigate()

	useEffect(() => {
		setAuthorsNotFetched(submission.authors)
	})

	// Get authors from their IDs.
	const setAuthorsNotFetched = (authors) => {
		authors.map((author) => {
			if (!authors.hasOwnProperty(author.userId)) {
				axiosInstance
					.get("/user/" + author.userId)
					.then((response) => {
						setAuthors((authors) => {
							return {
								...authors,
								[author.userId]: response.data
							}
						})
					})
					.catch((err) => {
						console.log(err)
					})
			}
		})
	}

	const getBadge = (submission) => {
		const [bg, text] = submission.approved
			? ["primary", "Approved"]
			: submission.approved === null
			? ["secondary", "In review"]
			: ["danger", "Rejected"]
		return (
			<Badge bg={bg} pill>
				{text}
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
		if (submission.hasOwnProperty("categories")) {
			let cards = submission.categories.map((category, i) => {
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
			return (
				<div
					style={{
						display: "flex",
						flexWrap: "wrap"
					}}>
					{cards}
				</div>
			)
		}
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
		<Card className={styles.SubmissionCard}>
			<Card.Body>
				<Card.Title>
					<div style={{ display: "flex" }}>
						<div style={{ flex: 1 }}>
							{cutShort(submission.name, 35)}
						</div>
						<div style={{ flex: 0.2, textAlign: "right" }}>
							{getBadge(submission)}
						</div>
					</div>
				</Card.Title>
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
			<Card.Body>{getTags(submission)}</Card.Body>
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
}
