import React, { useState, useEffect } from "react"
import styles from "./Submission.module.css"
import axiosInstance from "../../Web/axiosInstance"
import { useParams, useNavigate } from "react-router-dom"
import { CSSTransition, SwitchTransition } from "react-transition-group"
import FadeInTransition from "../../Components/Transitions/FadeIn.module.css"
import { Abstract, FileViewer } from "./Children"
import { Badge, Card, Button, Collapse } from "react-bootstrap"

function Submission() {
	// Router hooks
	const params = useParams()
	const navigate = useNavigate()

	// States for submission handling
	const [submission, setSubmission] = useState({
		ID: null,
		authors: [],
		reviewers: [],
		name: "Loading...",
		metaData: {
			abstract: "",
			reviews: []
		},
		approved: null
	})
	const [authors, setAuthors] = useState([])
	const [reviewers, setReviewers] = useState([])
	const [showError, setShowError] = useState(false)
	const [errMsg, setErrMsg] = useState("")
	const [showFile, setShowFile] = useState(false)
	const [fileId, setFileId] = useState(-1)

	useEffect(() => {
		if (!params.hasOwnProperty("id")) {
			navigate("/")
		}
		getSubmission(params.id)
	}, [])

	// Get the given submission from an ID.
	const getSubmission = (id) => {
		axiosInstance
			.get("/submission/" + id)
			.then((response) => {
				setSubmission(response.data)
				response.data.authors.map((author) =>
					getUser(author.userId, setAuthors)
				)
				if (response.data.hasOwnProperty("reviewers"))
					response.data.reviewers.map((reviewer) =>
						getUser(reviewer.userId, setReviewers)
					)
			})
			.catch((err) => {
				console.log(err)
				if (err.hasOwnProperty("response")) {
					if ([401, 404].includes(err.response.status)) {
						navigate("/")
					}
				}
			})
	}

	// Get an author's full profile from it's ID and add it to the authors array.
	const getUser = async (id, setUsers) => {
		try {
			let res = await axiosInstance.get("/user/" + id)
			setUsers((users) => {
				return [...users, res.data]
			})
		} catch (err) {
			console.log(err)
		}
	}

	// Get an author's full name.
	const getUserFullName = (author) => {
		if (author.hasOwnProperty("profile")) {
			return author.profile.firstName + " " + author.profile.lastName
		} else {
			return author.userId
		}
	}

	// Get status badge from submission status
	const getBadge = () => {
		let [bg, status] = submission.approved
			? ["primary", "Approved"]
			: submission.approved === null
			? ["secondary", "In review"]
			: ["danger", "Rejected"]
		return <Badge bg={bg}>{status}</Badge>
	}

	return (
		<div className={styles.SubmissionContainer}>
			<SwitchTransition>
				<CSSTransition
					key={!showFile}
					timeout={100}
					classNames={{ ...FadeInTransition }}>
					{showFile ? (
						<div style={{ display: "flex" }}>
							<p>{getBadge()}</p>
							<h5 style={{ marginLeft: "15px" }}>
								{submission.name}
							</h5>
						</div>
					) : (
						<div style={{ display: "flex" }}>
							<h2>{getBadge()}</h2>
							<h1 style={{ marginLeft: "15px" }}>
								{submission.name}
							</h1>
						</div>
					)}
				</CSSTransition>
			</SwitchTransition>
			<Collapse in={!showFile}>
				<div className="text-muted">
					<h5>
						Author
						{authors.length > 1 ? "s: " : ": "}
						{authors.length > 0
							? authors.map(
									(author, i) =>
										(i === 0 ? " " : ", ") +
										getUserFullName(author)
							  )
							: "No authors..."}
					</h5>
					<h5>
						Reviewer
						{reviewers.length > 1 ? "s: " : ": "}
						{reviewers.length > 0
							? reviewers.map((reviewer, i) =>
									i === 0 ? " " : ", "
							  )
							: "No reviewers..."}
					</h5>
				</div>
			</Collapse>
			<div style={{ display: "flex" }}>
				<div className={styles.LeftContainer}>
					<Abstract
						markdown={submission.metaData.abstract}
						show={showFile}
						setShow={(e) => setShowFile(e)}
						inversed
					/>
					<FileViewer id={fileId} show={showFile} />
				</div>
				<div className={styles.RightContainer}>There</div>
			</div>
		</div>
	)
}

export default Submission
