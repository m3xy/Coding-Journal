import React from "react"
import { Modal, Badge } from "react-bootstrap"
import ReactMarkdown from "react-markdown"

export default ({ review, reviewer, show, setShow }) => {
	const getFullName = (target) => {
		return target.profile.firstName + " " + target.profile.lastName
	}

	const getBadge = (approval) => {
		const [bg, text] = approval
			? ["primary", "Approves"]
			: ["danger", "Disapproves"]
		return <Badge bg={bg}>{text}</Badge>
	}

	if (reviewer !== undefined && review !== undefined)
		return (
			<Modal show={show} size="lg" onHide={() => setShow(false)}>
				<Modal.Header closeButton>
					<Modal.Title>
						<div style={{ display: "flex" }}>
							{getBadge(review.approved)}
							<div
								id="review-title"
								style={{ marginLeft: "15px" }}>
								Review by {getFullName(reviewer)}
							</div>
						</div>
					</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					<ReactMarkdown
						children={decodeURIComponent(
							escape(window.atob(review.base64Value))
						)}
					/>
				</Modal.Body>
			</Modal>
		)
	else return <div></div>
}
