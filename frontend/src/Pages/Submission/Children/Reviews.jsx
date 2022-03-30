/*
 * Review.jsx
 * Card listing all made reviews
 * Author: 190014935
 */
import React, { useState, useEffect } from "react"
import { Card, Badge, ListGroup, ListGroupItem } from "react-bootstrap"
import ReviewModal from "./ReviewModal"

export default ({ reviewers, reviews }) => {
	const [showModal, setShowModal] = useState(false)
	const [reviewerMap, setReviewerMap] = useState({})
	const [modalReview, setModalReview] = useState(<></>)

	useEffect(() => {
		if (Object.keys(reviewerMap).length > 0) setReviewerMap({})
		reviewers?.map((reviewer) => {
			setReviewerMap((reviewerMap) => {
				return { ...reviewerMap, [reviewer.userId]: reviewer }
			})
		})
	}, [reviewers])

	const getFullName = (reviewer) => {
		return reviewer.firstName + " " + reviewer.lastName
	}

	const getBadge = (approval) => {
		const [bg, text] = approval
			? ["primary", "Approves"]
			: ["danger", "Disapproves"]
		return <Badge bg={bg}>{text}</Badge>
	}

	const clickReview = (target) => {
		setModalReview(target)
		setShowModal(true)
	}

	return (
		<Card style={{ marginTop: "15px" }}>
			<Card.Body>
				<h4>Reviews</h4>
			</Card.Body>
			<ListGroup className="list-group-flush">
				{reviews?.map((review, i) => {
					if (reviewerMap[review.reviewerId] !== undefined)
						return (
							<ListGroupItem key={i}>
								<h5 style={{ display: "flex" }}>
									<Card.Link
										style={{ flex: "1" }}
										onClick={() => clickReview(review)}>
										{getFullName(
											reviewerMap[review.reviewerId]
										)}
									</Card.Link>
									<div
										style={{
											flex: "0.2",
											textAlign: "right"
										}}></div>
									{getBadge(review.approved)}
								</h5>
							</ListGroupItem>
						)
				})}
			</ListGroup>
			<ReviewModal
				show={showModal}
				setShow={setShowModal}
				review={modalReview}
				reviewer={reviewerMap[modalReview.reviewerId]}
			/>
		</Card>
	)
}
