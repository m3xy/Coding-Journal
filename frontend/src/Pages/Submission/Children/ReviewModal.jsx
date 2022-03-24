import React from "react"
import { Modal, Badge } from "react-bootstrap"
import ReactMarkdown from "react-markdown"

export default ({ review, reviewer, show, setShow }) => {
	const getFullName = (reviewer) => {
		return reviewer.profile.firstName + " " + reviewer.profile.lastName
	}

	const getBadge = (approval) => {
		const [bg, text] = approval
			? ["primary", "Approves"]
			: ["danger", "Disapproves"]
		return <Badge bg={bg}>{text}</Badge>
	}

	return (
		<Modal show={show} onHide={() => setShow(false)}>
			<Modal.Header closeButton>
				<Modal.Title>
					<div style={{ display: "flex" }}>
						<div id="review-title" style={{ flex: "0.2" }}>
							Review by {getFullName(reviewer)}
						</div>
						<div style={{ flex: "1", textAlign: "right" }}>
							{getBadge(review.approves)}
						</div>
					</div>
				</Modal.Title>
			</Modal.Header>
			<Modal.Body>
				<ReactMarkdown
					children={Buffer.from(
						review.base64Value,
						"bsae64"
					).toString("utf-8")}
				/>
			</Modal.Body>
		</Modal>
	)
}
