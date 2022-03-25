import React, { useState, useEffect } from "react"
import styles from "./Submission.module.css"
import axiosInstance from "../../Web/axiosInstance"
import { useParams, useNavigate } from "react-router-dom"
import { CSSTransition, SwitchTransition } from "react-transition-group"
import FadeInTransition from "../../Components/Transitions/FadeIn.module.css"
import JwtService from "../../Web/jwt.service"
import {
	Abstract,
	FileViewer,
	FileExplorer,
	TagsList,
	Reviews,
	ReviewEditor
} from "./Children"
import { Badge, Collapse, Button } from "react-bootstrap"

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
		files: [],
		approved: null
	})
	const [authors, setAuthors] = useState([])
	const [reviewers, setReviewers] = useState([])
	const [review, showReview] = useState(false)

	// Setters for file mode and file ID.
	const [showFile, setShowFile] = useState(false)
	const [fileId, setFileId] = useState(-1)

	// Error handling states
	const [showAlert, setAlert] = useState(false)
	const [alertMsg, setAlertMsg] = useState("")

	// Editor and reviewer variant controllers
	// 0 - user, 1 - author, 2 - reviewer, 3 - editor.
	const permissionLevel = ["editor", "author", "reviewer", "editor"]
	const [perm, setPermissions] = useState(0)

	useEffect(() => {
		// Check if the page has an ID
		if (!params.hasOwnProperty("id")) {
			navigate("/")
		}
		getSubmission(params.id)

		// Get required permissions
		setPermissions(() => {
			switch (JwtService.getUserType()) {
				case 3:
					return 2
				case 4:
					return 3
				default:
					return JwtService.getUserType()
			}
		})
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

	const getUsersString = (users, role) => {
		return (
			<h5>
				{role}
				{users.length > 1 ? "s: " : ": "}
				{users.length > 0
					? users.map(
							(user, i) =>
								(i === 0 ? " " : ", ") + getUserFullName(user)
					  )
					: "No " + role + "s..."}
			</h5>
		)
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
							<h2 style={{ flex: "0.1" }}>{getBadge()}</h2>
							<h1 style={{ flex: "1", marginLeft: "15px" }}>
								{submission.name}
							</h1>
							{permissionLevel[perm] === "editor" ? (
								<Button
									style={{
										flex: "0.15",
										justifyContent: "right"
									}}>
									Set Approval
								</Button>
							) : (
								permissionLevel[perm] === "reviewer" && (
									<Button
										style={{
											flex: "0.15",
											justifyContent: "right"
										}}
										onClick={() => showReview(true)}>
										Review
									</Button>
								)
							)}
						</div>
					)}
				</CSSTransition>
			</SwitchTransition>
			<Collapse in={!showFile}>
				<div className="text-muted">
					{getUsersString(authors, "Author")}
					<div style={{ display: "flex" }}>
						{permissionLevel[perm] === "editor" && (
							<Button style={{ marginRight: "5px" }} size="sm">
								Assign
							</Button>
						)}
						<div
							style={
								permissionLevel[perm] === "editor"
									? { marginTop: "5px" }
									: {}
							}>
							{getUsersString(reviewers, "Reviewer")}
						</div>
					</div>
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
				<div className={styles.RightContainer}>
					<FileExplorer
						files={submission.files}
						onClick={(id) => {
							setFileId(id)
							setShowFile(true)
						}}
					/>
					<TagsList tags={submission.categories} />
					{submission.metaData.hasOwnProperty("reviews") && (
						<Reviews
							reviews={submission.metaData.reviews}
							noProfileReviewers={
								submission.hasOwnProperty("reviewers")
									? submission.reviewers
									: []
							}
						/>
					)}
				</div>
			</div>
			<ReviewEditor
				id={params.id}
				show={review}
				setShow={showReview}
				setValidation={setAlert}
				setValidationMsg={setAlertMsg}
			/>
		</div>
	)
}

export default Submission
