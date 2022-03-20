import React, { useState, useEffect } from "react"
import { Card, Button } from "react-bootstrap"
import styles from "./SubmissionList.module.css"

export default ({ queryLambda, display }) => {
	const [submissions, setSubmissions] = useState([])
	const [error, setError] = useState(false)

	useEffect(() => {
		let { retSubs, retErr } = queryLambda()
		setSubmissions(retSubs)
		setError(retErr)
	})

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
								return (index === 0 ? " " : ", ") + author
							})}{" "}
						</Card.Subtitle>
					</Card.Body>
					<Card.Body
						style={{
							height: "60%",
							whiteSpace: "normal"
						}}>
						<Card.Text>
							{cutShort(submission.abstract, 200)}
						</Card.Text>
					</Card.Body>
					<Card.Body>
						<Button
							variant="primary"
							onClick={() => {
								navigate("/submissions/" + submission.id)
							}}>
							Explore
						</Button>
					</Card.Body>
					<Card.Footer className="text-muted">
						Created: {submission.createdAt}
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
				{error !== null
					? "Something went wrong, please try again later..."
					: submissions.length > 0
					? getCards(submissions)
					: "It's empty here..."}
			</div>
		</div>
	)
}
