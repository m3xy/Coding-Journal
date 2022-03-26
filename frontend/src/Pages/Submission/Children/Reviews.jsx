import React, { useEffect, useState } from "react"
import { Card, Badge, ListGroup, ListGroupItem } from "react-bootstrap"
import ReviewModal from "./ReviewModal"
import axiosInstance from "../../../Web/axiosInstance"

export default ({ noProfileReviewers, reviews }) => {
	const [reviewers, setReviewers] = useState({})
	const [showModal, setShowModal] = useState()
	const [modalReview, setModalReview] = useState(<></>)

	useEffect(() => {
		noProfileReviewers.map((reviewer) => {
			axiosInstance
				.get("/user/" + reviewer.userId)
				.then((response) => {
					setReviewers((reviewers) => {
						return {
							...reviewers,
							[reviewer.userId]: response.data
						}
					})
				})
				.catch((err) => {
					console.log(err)
				})
		})
	}, [noProfileReviewers])

	const getFullName = (reviewerId) => {
		if (reviewers.hasOwnProperty(reviewerId))
			return (
				reviewers[reviewerId].profile.firstName +
				" " +
				reviewers[reviewerId].profile.lastName
			)
		else return reviewerId
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
					return (
						<ListGroupItem key={i}>
							<h5 style={{ display: "flex" }}>
								<Card.Link
									style={{ flex: "1" }}
									onClick={() => clickReview(review)}>
									{getFullName(review.reviewerId)}
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
				reviewer={reviewers[modalReview.reviewerId]}
			/>
		</Card>
	)
}
