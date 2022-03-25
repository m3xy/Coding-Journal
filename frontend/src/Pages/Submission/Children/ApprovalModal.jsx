import React from "react"
import { Modal, Button, Table, ButtonGroup } from "react-bootstrap"

export default ({ submission, show, setShow }) => {
	return (
		<Modal show={show} onHide={() => setShow(false)} size="lg">
			<Modal.Header closeButton>
				<Modal.Title>Edit Submission</Modal.Title>
			</Modal.Header>
			<Modal.Body>
				<div className="text-muted">
					{submission.reviewers.length === 0
						? "No reviewers yet, please assign reviewers..."
						: (submission.metaData.hasOwnProperty('reviews') ? submission.metaData.reviews?.length : '0') +
						" reviewers have assigned a review, out of " + submission.reviewers?.length + " reviewers."}
				</div>
			</Modal.Body>
			<Modal.Body>
				<Table striped bordered hover>
					<thead>
						<th>#</th>
						<th>ID</th>
						<th>Approval</th>
					</thead>
					<tbody>
						{submission.metaData.reviews?.map((review, i) => (
							<tr key={i}>
								<td>{i}</td>
								<td>{review.reviewerId}</td>
								<td>{review.approved ? "Approved" : "Disapproved"}</td>
							</tr>
						))}
					</tbody>
				</Table>
			</Modal.Body>
			<Modal.Footer>
				<ButtonGroup className="mb-2">
					<Button disabled={!submission.metaData.hasOwnProperty('reviews')}>Approve</Button>
					<Button variant="danger">Reject</Button>
				</ButtonGroup>
				<Button className="mb-2">Migrate</Button>
			</Modal.Footer>
		</Modal>
	)
}
