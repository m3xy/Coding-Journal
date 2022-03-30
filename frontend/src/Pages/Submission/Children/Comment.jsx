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
import { useNavigate } from "react-router-dom"
import JwtService from "../../../Web/jwt.service"
import axiosInstance from "../../../Web/axiosInstance"
import ReactMarkdown from "react-markdown"

const userEndpoint = "/user"
const editEndpoint = "/edit"
const deleteEndpoint = "/delete"
const profileURL = "/profile"

function Comment({
	fileID,
	comment,
	show,
	postReply,
	fileEndpoint,
	commentEndpoint,
	refresh
}) {
	//Used to navigate between routes of our react application
	const navigate = useNavigate()

	//State variable storing the reply text to the comment
	const [reply, setReply] = useState("")

	//State variable which determines whether replies for the comment are shown or not
	const [showReplies, setShowReplies] = useState(false)

	//State variable which determines whether the reply box is visible or not
	const [openReply, setOpenReply] = useState(false)

	//State variable, the author name of the comment
	const [name, setName] = useState("")

	//State variable storing the text of a comment edit
	const [edit, setEdit] = useState("")

	//State variable which determines whether the edit box is visible or not
	const [openEdit, setOpenEdit] = useState(false)

	//Fetch the comment author's name when the comment component has (re)rendered
	useEffect(() => {
		getName()
	}, [name])

	//Fetch the comment author's name
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

	//Makes a post request including the edits made to the comment
	const postEdit = () => {
		const b64Edit = btoa(edit)
		if (b64Edit == comment.base64Value) {
			console.log("Comment unchanged")
		}

		let data = {
			base64Value: b64Edit
		}

		axiosInstance
			.post(
				fileEndpoint +
					"/" +
					fileID +
					commentEndpoint +
					"/" +
					comment.ID +
					editEndpoint,
				data
			)
			.then((response) => {
				console.log(response)
				refresh()
				setOpenEdit(false)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	//Makes a post request to delete the comment
	const postDelete = () => {
		axiosInstance
			.post(
				fileEndpoint +
					"/" +
					fileID +
					commentEndpoint +
					"/" +
					comment.ID +
					deleteEndpoint
			)
			.then((response) => {
				console.log(response)
				refresh()
			})
			.catch((error) => {
				console.log(error)
			})
	}

	//Replies are represented as further comment components displayed under their parent comment
	const repliesHTML = comment.comments
		? comment.comments.map((reply, i) => {
				return (
					<Comment
						key={i}
						fileID={fileID}
						comment={reply}
						show={showReplies}
						postReply={postReply}
						fileEndpoint={fileEndpoint}
						commentEndpoint={commentEndpoint}
						refresh={refresh}
					/>
				)
		  })
		: ""

	return (
		show && (
			<Toast
				style={{ verticalAlign: "top", minWidth: "100%" }}
				className="d-inline-block m-1">
				<Toast.Header closeButton={false}>
					<Button
						variant="light"
						className="me-auto"
						size="sm"
						onClick={() =>
							navigate(profileURL + "/" + comment.author)
						}>
						{name}
					</Button>
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
					{openEdit ? (
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
						<ReactMarkdown>
							{atob(comment.base64Value)}
						</ReactMarkdown>
					)}
					<p />
					<small className="text-muted">
						{comment.DeletedAt
							? "(deleted)"
							: comment.CreatedAt == comment.UpdatedAt
							? new Date(comment.CreatedAt).toDateString()
							: new Date(comment.UpdatedAt).toDateString() +
							  " (edited)"}
					</small>
					<br />

					<Button
						variant="light"
						onClick={() => setOpenReply(!openReply)}>
						↩
					</Button>
					{JwtService.getUserID() == comment.author && (
						<>
							<Button
								variant="light"
								onClick={() => {
									setEdit(atob(comment.base64Value))
									setOpenEdit(!openEdit)
								}}>
								✎
							</Button>
							<Button variant="light" onClick={postDelete}>
								🗑
							</Button>
						</>
					)}
					{comment.comments && (
						<Button
							variant="light"
							onClick={() => setShowReplies(!showReplies)}>
							💬
						</Button>
					)}

					<Collapse in={openReply}>
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
									setOpenReply(false)
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
