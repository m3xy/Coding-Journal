/**
 * Comment.jsx
 * Author: 190019931
 *
 * React component for displaying a comment
 */

import React, { useEffect, useState } from "react"
import {
	Button,
	Collapse,
	FormControl,
	InputGroup,
	Toast
} from "react-bootstrap"
import JwtService from "../Web/jwt.service"
import axiosInstance from "../Web/axiosInstance"

const userEndpoint = "/user"
const editEndpoint = "/edit"
const deleteEndpoint = "/delete"

function Comment({ fileID, comment, show, postReply, fileEndpoint, commentEndpoint }) {
	const [reply, setReply] = useState("")
	const [showReplies, setShowReplies] = useState(false)
	const [openReplies, setOpenReplies] = useState(false)
	const [name, setName] = useState("")
	const [edit, setEdit] = useState("")
	const [openEdits, setOpenEdits] = useState(false)

	useEffect(() => {
		getName()
	}, [name])

	const getName = () => {
		axiosInstance
			.get(userEndpoint + "/" + comment.author)
			.then((response) => {
				setName(response.data.firstName + " " + response.data.lastName)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const postEdit = () => {
        const b64Edit = btoa(edit)
		if (b64Edit == comment.base64Value) {
			console.log("Comment unchanged")
		}

        let data = {
            base64Value: b64Edit
        }
        
        axiosInstance
            .post(fileEndpoint + "/" + fileID + commentEndpoint + "/" + comment.ID + editEndpoint, data)
            .then((response) => {
                console.log(response)
            }).catch((error) => {
                console.log(error)
            })
	}

    const postDelete = () => {
        axiosInstance
            .post(fileEndpoint + "/" + fileID + commentEndpoint + "/" + comment.ID + deleteEndpoint)
            .then((response) => {
                console.log(response)
            }).catch((error) => {
                console.log(error)
            })
    }

	const repliesHTML = comment.comments
		? comment.comments.map((reply) => {
				return (
					<Comment
						comment={reply}
						show={showReplies}
						postReply={postReply}
					/>
				)
		  })
		: ""

	return (
		show && (
			<Toast
				style={{ verticalAlign: "top" }}
				className="d-inline-block m-1">
				<Toast.Header closeButton={false}>
					<strong className="me-auto">{name}</strong>
					<small>
						{comment.startLine == comment.endLine ? (
							<>Line: {comment.startLine}</>
						) : (
							<>
								Lines: {comment.startLine} - {comment.endLine}
							</>
						)}
					</small>
				</Toast.Header>
				<Toast.Body>
					{openEdits ? (
						<InputGroup className="mb-3" size="sm">
							<FormControl
								placeholder={"Edit comment"}
								onChange={(e) => setEdit(e.target.value)}
								value={edit}
							/>
							<Button
								variant="outline-secondary"
								onClick={(e) => {
									postEdit()
									setEdit("")
								}}>
								Save
							</Button>
						</InputGroup>
					) : (
						atob(comment.base64Value)
					)}
					<p />
					<small className="text-muted">
						{comment.DeletedAt
							? "(deleted)"
							: comment.CreatedAt == comment.UpdatedAt
							? comment.CreatedAt
							: comment.UpdatedAt + " (edited)"}
					</small>
					<br />

					<Button
						variant="light"
						onClick={() => setOpenReplies(!openReplies)}>
						↩
					</Button>
					{JwtService.getUserID() == comment.author && (
						<>
							<Button
								variant="light"
								onClick={() => {
									setEdit(atob(comment.base64Value))
									setOpenEdits(!openEdits)
								}}>
								✎
							</Button>
							<Button variant="light" onClick={postDelete}>🗑</Button>
						</>
					)}
					{comment.comments && (
						<Button
							variant="light"
							onClick={() => setShowReplies(!showReplies)}>
							💬
						</Button>
					)}

					<Collapse in={openReplies}>
						<InputGroup className="mb-3" size="sm">
							<FormControl
								placeholder={"Enter a reply"}
								onChange={(e) => setReply(e.target.value)}
								value={reply}
							/>
							<Button
								variant="outline-secondary"
								onClick={(e) => {
									postReply(e, comment.ID, reply)
									setReply("")
								}}>
								Reply
							</Button>
						</InputGroup>
					</Collapse>
				</Toast.Body>
				{repliesHTML}
			</Toast>
		)
	)
}

export default Comment
