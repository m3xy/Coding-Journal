import React, { useEffect, useState } from "react"
import { Card, Badge, ListGroup, ListGroupItem } from "react-bootstrap"
import axiosInstance from "../../../Web/axiosInstance"

export default ({ reviewerIds, reviews }) => {
	const [reviewers, setReviewers] = useState({})

	useEffect(() => {
		reviewerIds.map((reviewerId) => {
			axiosInstance.get('/user/' + reviewerId.userId)
				.then((response) => {
					setReviewers(reviewers => { return { ...reviewers, [reviewerId.userId]: response.data } })
				})
				.catch((err) => {
					console.log(err)
				})
		})
	}, [])

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
		<Card body>
			<h4>Reviews</h4>
			<ListGroup>
				{reviews.map((review) => {
					<ListGroupItem>
						<div style={{ display: "flex" }}>
							<Card.Link style={{ flex: "0.2" }}>
								{getFullName(reviewers.hasOwnProperty(review.reviewerId) ?
									reviewers[review.reviewerId] :
									review.reviewerId)}
							</Card.Link>
							<div style={{ flex: "1" }}>
							</div>
							{getBadge(review.approved)}
						</div>
					</ListGroupItem>
				})}
			</ListGroup>
		</Card>
	)
}
