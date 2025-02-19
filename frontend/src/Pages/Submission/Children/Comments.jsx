/**
 * Comments.jsx
 * Author: 190019931
 *
 * React component for displaying comments
 */

import React, { useState } from "react"
import { Button, Form, Modal } from "react-bootstrap"
import axiosInstance from "../../../Web/axiosInstance"
import JwtService from "../../../Web/jwt.service"
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
	//State variable storing the text content of the comments box
	const [text, setText] = useState("")

	//State variable describing whether or not comments are shown
	const [showComment, setShowComment] = useState(true)

	//State variable storing whether the comments are loading (being fetched)
	const [isLoading, setLoading] = useState(false)

	//Function for posting a comment - parameters include the parent comment ID, content of the comment, and the start/end lines, span of the comment
	const postComment = (e, parentID, content, startLine, endLine) => {
		e.preventDefault()

		if (!content || content == "") {
			console.log("No comment written")
			return
		}

		let userId = JwtService.getUserID()
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

	//load more comments (if available) by refetching the file
	const loadMore = () => {
		setLoading(true)
		refresh()
		setLoading(false)
	}

	//HTML for each of the (parent) comments (Map each comment to a comment component to display)
	const commentsHTML =
		comments !== undefined
			? comments.map((comment, i) => {
					return (
						<Comment
							key={i}
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
							refresh={refresh}
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
						<Button
							variant="dark"
							type="submit"
							style={{ float: "right" }}>
							Post comment
						</Button>
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
					<br />
					<div className="d-grid gap-2">
						{comments?.length > 0 ? (
							commentsHTML
						) : (
							<small
								className="text-muted"
								style={{ textAlign: "center" }}>
								No comments
							</small>
						)}
						<Button
							variant="light"
							disabled={isLoading}
							onClick={!isLoading ? () => loadMore() : null}>
							{isLoading ? "…" : "↻"}
						</Button>
					</div>
				</Modal.Body>
				<Modal.Footer>
					<small className="text-muted">
						Last refreshed: {new Date().toLocaleString()}
					</small>
				</Modal.Footer>
			</Form>
		</Modal>
	)
}

export default Comments
