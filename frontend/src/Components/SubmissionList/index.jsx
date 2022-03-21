import React, { useState, useEffect } from "react"
import { Card, Button } from "react-bootstrap"
import styles from "./SubmissionList.module.css"
import { useNavigate } from "react-router-dom"
import axiosInstance from "../../Web/axiosInstance"

export default ({ query, display }) => {
	const [submissions, setSubmissions] = useState([])
	const [error, setError] = useState(null)
	const navigate = useNavigate()

	useEffect(() => {
		axiosInstance
			.get("/submissions/query", {
				params: query
			})
			.then((response) => {
				response.data.submissions.map((submission) => {
					axiosInstance
						.get("/submission/" + submission.ID)
						.then((response) => {
							setSubmissions((submissions) => {
								return [...submissions, response.data]
							})
						})
						.catch((err) => {
							console.log(err)
						})
				})
			})
			.catch(() => {
				setError(true)
			})
	}, [])

	const getCards = (submissions) => {
		return submissions.map((submission) => {
			return (
				<Card
					key={submission.ID}
					style={{
						minWidth: "18rem",
						margin: "8px"
					}}
					className="shadow rounded">
					<Card.Body>
						<Card.Title>{cutShort(submission.name, 40)}</Card.Title>
						<Card.Subtitle className="mb-2 text-muted">
							{" "}
							{submission.authors.length > 1
								? "Authors:"
								: "Author:"}{" "}
							{submission.authors.map((author, index) => {
								return (
									(index === 0 ? " " : ", ") + author.fullName
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
						Created: {submission.CreatedAt}
					</Card.Footer>
				</Card>
			)
		})
	}

	// Shorten a string if it is too long for given format.
	const cutShort = (text, limit) => {
		if (text.length > limit) {
			let short = text.substring(0, limit)
			return short.substring(0, short.lastIndexOf(" ")) + "..."
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
