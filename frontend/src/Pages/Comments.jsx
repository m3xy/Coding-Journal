/**
 * Comments.jsx
 * Author: 190019931
 *
 * React component for displaying comments
 */

import React, { useEffect, useState } from "react"
import { Button, Form, Modal, Toast } from "react-bootstrap"
import axiosInstance from "../Web/axiosInstance"
import JwtService from "../Web/jwt.service"
import Comment from "./Comment"

const fileEndpoint = "/file"
const commentEndpoint = "/comment"

function Comments({
	id,
	comments,
	startLine,
	endLine,
	show,
	setShow,
	refresh
}) {
	const [text, setText] = useState("")
	const [showComment, setShowComment] = useState(true)
	const [isLoading, setLoading] = useState(false)

	const postComment = (e, parentID, content, startLine, endLine) => {
		e.preventDefault()

		if (!content || content == "") {
			console.log("No comment written")
			return
		}

		let userId = JwtService.getUserID() //Preparing to get userId from session cookie
		if (!userId) {
			//If user has not logged in, disallow submit
			console.log("Not logged in")
			return
		}

		let comment = {
			parentId: parentID,
			base64Value: btoa(content),
			startLine: startLine,
			endLine: endLine,
			author: userId
		}
		console.log(comment)
		axiosInstance
			.post(fileEndpoint + "/" + id + commentEndpoint, comment)
			.then((response) => {
				refresh()
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const loadMore = () => {
		setLoading(true)
		setTimeout(() => {
			setLoading(false)
		}, 1000)
	}

	const commentsHTML =
		comments !== undefined
			? comments.map((comment) => {
					return (
						<Comment
							fileID={id}
							comment={comment}
							show={showComment}
							postReply={(e, parentID, content) => {
								postComment(
									e,
									parentID,
									content,
									comment.startLine,
									comment.endLine
								)
							}}
							fileEndpoint={fileEndpoint}
							commentEndpoint={commentEndpoint}
						/>
					)
			  })
			: ""

	return (
		<Modal show={show} onHide={() => setShow(false)} size="lg">
			<Form
				onSubmit={(e) => {
					setText("")
					postComment(e, null, text, startLine, endLine)
				}}>
				<Modal.Header closeButton>
					<Modal.Title>Comments</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					{commentsHTML}
					<div className="d-grid gap-2">
						<Button
							variant="link"
							disabled={isLoading}
							onClick={!isLoading ? () => loadMore() : null}>
							{isLoading ? "Loading…" : "Load more"}
						</Button>
					</div>
					<br />
					<Form.Group className="mb-3" controlId="CommentText">
						<Form.Label>Enter a comment below:</Form.Label>
						<Form.Control
							as="textarea"
							rows={3}
							aria-describedby="lineNumber"
							onChange={(e) => {
								setText(e.target.value)
							}}
							value={text}
						/>
						<Form.Text id="lineNumber" muted>
							{startLine == endLine ? (
								<>Line: {startLine}</>
							) : (
								<>
									Lines: {startLine} - {endLine}
								</>
							)}
						</Form.Text>
					</Form.Group>
				</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={() => setShow(false)}>
						Close
					</Button>
					<Button variant="primary" type="submit">
						Post comment
					</Button>
				</Modal.Footer>
			</Form>
		</Modal>
	)
}

export default Comments
