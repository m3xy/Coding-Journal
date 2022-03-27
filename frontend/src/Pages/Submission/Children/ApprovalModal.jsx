import React, { useState } from "react"
import { Modal, Button, Table, ButtonGroup, Alert } from "react-bootstrap"
import { CSSTransition } from "react-transition-group"
import FadeInStyles from "../../../Components/Transitions/FadeIn.module.css"
import axiosInstance from "../../../Web/axiosInstance"

export default ({ submission, show, setShow, setAlertMsg, showAlertMsg }) => {
	const [showAlert, setShowAlert] = useState(false)
	const [msg, setMsg] = useState("")

	const submitApproval = (approval) => {
		axiosInstance
			.post("/submission/" + submission.ID + "/approve", {
				status: approval
			})
			.then((response) => {
				setAlertMsg(response.message)
				showAlertMsg(true)
				setShow(false)
			})
			.catch((err) => {
				console.log(err)
				setMsg("Error - " + err.response?.data.message)
				setShowAlert(false)
			})
	}

	return (
		<Modal show={show} onHide={() => setShow(false)} size="lg">
			<Modal.Header closeButton>
				<Modal.Title>Edit Submission</Modal.Title>
			</Modal.Header>
			<Modal.Body>
				<div className="text-muted">
					{submission.reviewers === undefined
						? "No reviewers yet, please assign reviewers..."
						: (submission.metaData.hasOwnProperty("reviews")
								? submission.metaData.reviews?.length
								: "0") +
						  " reviewers have assigned a review, out of " +
						  submission.reviewers?.length +
						  " reviewers."}
				</div>
			</Modal.Body>
			<Modal.Body>
				<Table striped bordered hover>
					<thead>
						<tr>
							<th>#</th>
							<th>ID</th>
							<th>Approval</th>
						</tr>
					</thead>
					<tbody>
						{submission.metaData.reviews?.map((review, i) => (
							<tr key={i}>
								<td>{i}</td>
								<td>{review.reviewerId}</td>
								<td>
									{review.approved
										? "Approved"
										: "Disapproved"}
								</td>
							</tr>
						))}
					</tbody>
				</Table>
			</Modal.Body>
			<CSSTransition
				in={showAlert}
				timeout={100}
				unmountOnExit
				classNames={{ ...FadeInStyles }}>
				<Modal.Body>
					<Alert
						variant="danger"
						onClose={() => setShowAlert(false)}
						dismissible>
						{msg}
					</Alert>
				</Modal.Body>
			</CSSTransition>
			<Modal.Footer>
				<ButtonGroup className="mb-2">
					<Button
						onClick={() => submitApproval(true)}
						disabled={submission.metaData.reviews === undefined}>
						Approve
					</Button>
					<Button
						variant="danger"
						onClick={() => submitApproval(false)}>
						Reject
					</Button>
				</ButtonGroup>
				<Button className="mb-2">Migrate</Button>
			</Modal.Footer>
		</Modal>
	)
}
