import axios from "axios"
import React, {useState, useEffect} from "react"
import axiosInstance from "../../Web/axiosInstance"
import styles from "./Card.module.css"

export default function Card({submissionID}) {

	const [loaded, setLoaded] = useState(false)
	const [submission, setSubmission] = useState(null)

	useEffect(() => {
		// Get latest submissions
		axiosInstance.get("/submission/" + submissionID.toString())
		.then((response) => {
			// Get response as an array
			setSubmission(response.data)
			setLoaded(true)
		})
	}, [submissionID, setSubmission])

	// Shorten a string if it is too long for given format.
	function cutShort(text, limit) {
		if (text.length > limit) {
			let short = text.substring(0, limit)
			return short.substring(0, short.lastIndexOf(" ")) + "..."
		} else return text
	}

	return(
		<div className={styles.Card}>
			<div className={styles.TitleHalf}>
				<div className={styles.marginedParagraph}>
					{loaded? cutShort(submission.name, 40): ""}
				</div>
			</div>
			<div className={styles.AbstractHalf}>
				<p className={styles.marginedParagraph}>
					{loaded? cutShort(submission.abstract, 1024): ""}
				</p>
			</div>
		</div>
	)
}
