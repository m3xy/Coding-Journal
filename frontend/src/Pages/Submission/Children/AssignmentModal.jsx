/*
 * AssignmentModal.jsx
 * Modal for reviewer assignment
 * Author: 190014935
 */
import React, { useState } from "react"
import { Modal, Button } from "react-bootstrap"
import { FormUser } from "../../../Components/FormComponents"
import axiosInstance from "../../../Web/axiosInstance"

export default ({
	reviewers,
	submissionID,
	show,
	setShow,
	setAlertMsg,
	showAlertMsg
}) => {
	const [newReviewers, setNewReviewers] = useState([])

	const assignReviewers = () => {
		const data = {
			reviewers: [
				...newReviewers.map((reviewer) => {
					return reviewer.userId
				})
			]
		}
		axiosInstance
			.post("/submission/" + submissionID + "/assignreviewers", data)
			.then(() => {
				setAlertMsg(
					"Reviewer assignment successful! Please reload page..."
				)
				showAlertMsg(true)
				setShow(false)
			})
			.catch((err) => {
				console.log(err)
			})
	}

	return (
		<Modal show={show} onHide={() => setShow(false)} size="lg">
			<Modal.Header closeButton> Assign Reviewers </Modal.Header>
			<Modal.Body>
				<FormUser
					display="Reviewers"
					name="reviewers"
					immutables={reviewers}
					query={{ userType: 2 }}
					onChange={(e) => {
						setNewReviewers(e.target.value)
					}}
				/>
				<Button
					onClick={() => {
						assignReviewers()
					}}>
					Proceed
				</Button>
			</Modal.Body>
		</Modal>
	)
}
