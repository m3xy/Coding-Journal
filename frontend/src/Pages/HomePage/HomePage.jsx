import React, {useState, useEffect} from "react"
import axiosInstance from "../../Web/axiosInstance"
import styles from "./HomePage.module.css"
import { Card, Row, Button } from "react-bootstrap"
import { useNavigate } from "react-router-dom"

export default () => {
    const [submissions, setSubmissions] = useState([])
    const navigate = useNavigate()

    useEffect(() => {
        axiosInstance.get("/submissions", {params: { timestamp: new Date(Date.now()).toISOString()}})
        .then((response) => {
            let submissionIDs = []
            response.data.submissions.map((submission) => {
                submissionIDs.push(submission)
            })
            setSubmissions(submissionIDs)
        })
        .catch((err) => {
        })
    }, [])

	// Shorten a string if it is too long for given format.
	function cutShort(text, limit) {
		if (text.length > limit) {
			let short = text.substring(0, limit)
			return short.substring(0, short.lastIndexOf(" ")) + "..."
		} else return text
	}

    return(
    <div className={styles.HomePage}>
        <div className={styles.HomeContent}>
            <div className={styles.scrollerComponent}>
                <h2>
                    Most Recent Submissions
                </h2>
                <div className={styles.ScrollerContainer}>
                        {submissions.length > 0 ? submissions.map((submission) => {
                            return(
                                <Card key={submission.id} style={{minWidth: '18rem', margin: '8px'}} className="shadow rounded">
                                    <Card.Body>
                                        <Card.Title>
                                            { cutShort(submission.name, 40) }
                                        </Card.Title>
                                        <Card.Subtitle className="mb-2 text-muted"> {submission.authors.length > 1 ? "Authors:": "Author:"} {submission.authors.map((author, index) => {
                                            return (index === 0? " ": ", ") + author
                                        })} </Card.Subtitle>
                                    </Card.Body>
                                    <Card.Body style={{ height: '60%', whiteSpace: 'normal'}}>
                                        <Card.Text>
                                                { cutShort(submission.abstract, 200)}
                                        </Card.Text>
                                    </Card.Body>
                                    <Card.Body>
                                        <Button variant="primary" onClick={() => {navigate("/submissions/" + submission.id)}}>Explore</Button>
                                    </Card.Body>
                                    <Card.Footer className="text-muted">
                                         Created: {submission.createdAt}
                                    </Card.Footer>
                                </Card>
                            )
                        }): <div>No submissions available.</div>}
                </div>
            </div>
        </div>
    </div>
    )
}