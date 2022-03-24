import React, { useState, useEffect } from "react"
import { Card, Button, Badge } from "react-bootstrap"
import styles from "./SubmissionList.module.css"
import SubmissionCard from "../SubmissionCard"
import axiosInstance from "../../Web/axiosInstance"

export default ({ query, display }) => {
	const [submissions, setSubmissions] = useState([])
	const [error, setError] = useState(null)

	useEffect(() => {
		axiosInstance
			.get("/submissions/query", {
				params: query
			})
			.then((response) => {
				if (response.data.hasOwnProperty("submissions"))
					setSubmissionsFromPrimitives(response.data.submissions)
			})
			.catch((err) => {
				console.log(err)
				setError(true)
			})
	}, [])

	// Get submission details from their basic types.
	const setSubmissionsFromPrimitives = (submissionPrimitives) => {
		if (submissionPrimitives !== null)
			submissionPrimitives.map((submission) => {
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
	}

	// Get a list of cards from the list of submissions.
	const getCards = (submissions) => {
		return submissions.map((submission, i) => {
			return <SubmissionCard key={i} submission={submission} />
		})
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
