/*
 * Submission.jsx
 * Main file for the submission page.
 * Author: 190014935
 */

import React, { useState, useEffect } from "react"
import styles from "./Submission.module.css"
import axiosInstance from "../../Web/axiosInstance"
import { useParams, useNavigate } from "react-router-dom"
import { CSSTransition, SwitchTransition } from "react-transition-group"
import FadeInTransition from "../../Components/Transitions/FadeIn.module.css"
import JwtService from "../../Web/jwt.service"
import Compiler from "../../Components/Compiler"
import {
	Abstract,
	FileViewer,
	FileExplorer,
	TagsList,
	Reviews,
	ReviewEditor as EditorModal,
	AssignmentModal,
	ApprovalModal
} from "./Children"
import {
	Alert,
	Badge,
	Collapse,
	Button,
	ButtonGroup,
	DropdownButton,
	Dropdown
} from "react-bootstrap"

const OTHER_JOURNALS = [2, 5, 8, 13, 17, 20, 23, 26]

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
		files: [{ ID: null, path: "" }],
		approved: null
	})
	const [review, showReview] = useState(false)
	const [approval, showApproval] = useState(false)
	const [assignment, showAssignment] = useState(false)

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

	// Get an author's full name.
	const getUserFullName = (author) => {
		return author?.firstName + " " + author?.lastName
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

	// Get list of users in human-readable format( user: name1, name2, ... )
	const getUsersString = (users, role) => {
		return (
			<h5>
				{role}
				{users?.length > 1 ? "s: " : ": "}
				{users?.length > 0
					? users?.map(
							(user, i) =>
								(i === 0 ? " " : ", ") + getUserFullName(user)
					  )
					: "No " + role + "s..."}
			</h5>
		)
	}

	// Export the submission to another journal
	const exportSubmission = (journal) => {
		axiosInstance
			.post("/submission/" + submission.ID + "/export/" + journal)
			.then(() => {
				setAlertMsg("Export successful")
				setAlert(true)
			})
			.catch((error) => console.log(error))
	}

	// Buttons for editor and reviewer, for review posting
	// and submission approval.
	const permissionButtons = () => {
		if (permissionLevel[perm] === "editor")
			return (
				<ButtonGroup vertical>
					<Button
						onClick={() => showApproval(true)}
						style={{
							flex: "0.15",
							justifyContent: "right"
						}}>
						Set Approval
					</Button>
					<DropdownButton
						as={ButtonGroup}
						title="Export submission"
						variant="outline-secondary">
						{OTHER_JOURNALS.map((journal) => {
							return (
								<Dropdown.Item
									onClick={() => exportSubmission(journal)}>
									Journal {journal}
								</Dropdown.Item>
							)
						})}
					</DropdownButton>
				</ButtonGroup>
			)
		else if (
			permissionLevel[perm] === "reviewer" &&
			submission.reviewers
				?.map((reviewer) => {
					return reviewer.userId
				})
				.includes(JwtService.getUserID())
		) {
			const disabled = submission.metaData.reviews
				?.map((review) => {
					return review.reviewerId
				})
				.includes(JwtService.getUserID())
			if (!disabled)
				return (
					<Button
						style={{
							flex: "0.15",
							justifyContent: "right"
						}}
						onClick={
							!disabled ? () => showReview(true) : () => null
						}>
						Review
					</Button>
				)
			else
				return (
					<h5 className={`text-muted ${styles.DisabledReviewButton}`}>
						You've already reviewed this submission...
					</h5>
				)
		}
	}

	// Container shown on the left side of the page.
	const leftContainer = () => {
		return (
			<div className={styles.LeftContainer}>
				<Abstract
					markdown={submission.metaData.abstract}
					show={showFile}
					setShow={(e) => setShowFile(e)}
					inversed
				/>
				<FileViewer id={fileId} show={showFile} />
			</div>
		)
	}

	// Container shown on the right side of the page.
	const rightContainer = () => {
		return (
			<div className={styles.RightContainer}>
				<FileExplorer
					files={submission.files}
					onClick={(id) => {
						setShowFile(true)
						setFileId(id)
					}}
				/>
				<TagsList tags={submission.categories} />
				<Compiler id={params.id} submission={submission} />
				{submission.metaData.hasOwnProperty("reviews") && (
					<Reviews
						reviews={submission.metaData.reviews}
						reviewers={
							submission.hasOwnProperty("reviewers")
								? submission.reviewers
								: []
						}
					/>
				)}
				
			</div>
		)
	}

	return (
		<div className={styles.SubmissionContainer}>
			<CSSTransition
				in={showAlert}
				timeout={100}
				unmountOnExit
				classNames={{ ...FadeInTransition }}>
				<Alert
					variant="success"
					dismissible
					onClose={() => {
						setAlert(false)
					}}>
					{alertMsg}
				</Alert>
			</CSSTransition>
			<SwitchTransition>
				<CSSTransition
					key={!showFile}
					timeout={100}
					classNames={{ ...FadeInTransition }}>
					{/* Transition on the page's header - big for abstract, small for files. */}
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
							{permissionButtons()}
						</div>
					)}
				</CSSTransition>
			</SwitchTransition>
			{/* Authors and reviewers - if editor, show assignment button on reviewers */}
			<Collapse in={!showFile}>
				<div className="text-muted">
					{getUsersString(submission.authors, "Author")}
					<div style={{ display: "flex" }}>
						{permissionLevel[perm] === "editor" && (
							<Button
								style={{ marginRight: "5px" }}
								size="sm"
								onClick={() => showAssignment(true)}>
								Assign
							</Button>
						)}
						<div
							className={
								permissionLevel[perm] === "editor"
									? styles.ButtonTextCenter
									: ""
							}>
							{getUsersString(submission.reviewers, "Reviewer")}
						</div>
					</div>
				</div>
			</Collapse>
			{/* Main body - contains left and right containers. */}
			<div style={{ display: "flex" }}>
				{leftContainer()}
				{rightContainer()}
			</div>
			<EditorModal
				id={params.id}
				show={review}
				setShow={showReview}
				setValidation={setAlert}
				setValidationMsg={setAlertMsg}
			/>
			<AssignmentModal
				reviewers={submission.reviewers}
				submissionID={submission.ID}
				show={assignment}
				setShow={showAssignment}
				showAlertMsg={setAlert}
				setAlertMsg={setAlertMsg}
			/>
			<ApprovalModal
				submission={submission}
				show={approval}
				setShow={showApproval}
				showAlertMsg={setAlert}
				setAlertMsg={setAlertMsg}
			/>
		</div>
	)
}

export default Submission
