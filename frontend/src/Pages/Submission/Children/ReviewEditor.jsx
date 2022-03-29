/*
 * ReviewEditor.jsx
 * Modal for review submissions
 * Author: 190014935
 */
import React, { useEffect, useState } from "react"
import { Editor } from "react-draft-wysiwyg"
import { draftToMarkdown } from "markdown-draft-js"
import { EditorState, convertToRaw } from "draft-js"
import editorStyles from "./Submission.module.css"
import "react-draft-wysiwyg/dist/react-draft-wysiwyg.css"
import { Modal, Form, Alert, Button, Tabs, Tab } from "react-bootstrap"
import ReactMarkdown from "react-markdown"
import axiosInstance from "../../../Web/axiosInstance"

export default ({ id, show, setShow, setValidation, setValidationMsg }) => {
	const [markdown, setMarkdown] = useState("")
	const [editorState, setEditorState] = useState(() =>
		EditorState.createEmpty()
	)
	const [approve, setApproval] = useState(false)
	const [showErr, setShowErr] = useState(false)
	const [errMsg, setErrMsg] = useState("")

	useEffect(() => {
		setMarkdown(
			draftToMarkdown(convertToRaw(editorState.getCurrentContent()))
		)
	}, [editorState])

	const uploadReview = () => {
		console.log(markdown)
		axiosInstance
			.post("/submission/" + id + "/review", {
				approved: approve,
				base64Value: window.btoa(unescape(encodeURIComponent(markdown)))
			})
			.then(() => {
				setValidationMsg("Review successfully posted!")
				setValidation(true)
				setShow(false)
			})
			.catch((error) => {
				console.log(error)
				setErrMsg(
					"Review upload failed - " +
						error.response.data.message +
						" - " +
						error.response.status
				)
				setShowErr(true)
			})
	}

	useEffect(() => {}, [editorState])

	return (
		<Modal size="lg" show={show} onHide={() => setShow(false)}>
			<Modal.Header closeButton>
				<h3>
					<div>Review</div>
				</h3>
			</Modal.Header>
			<Modal.Body>
				<Tabs defaultActiveKey="edit">
					<Tab eventKey="edit" title="Edit">
						<div className={editorStyles.editor}>
							<Editor
								editorState={editorState}
								toolbarClassName="toolbarClassName"
								wrapperClassname="wrapperClassName"
								editorClassName="editorClassName"
								onEditorStateChange={setEditorState}
							/>
						</div>
					</Tab>
					<Tab eventKey="preview" title="Preview">
						<div className={editorStyles.editor}>
							<ReactMarkdown children={markdown} />
						</div>
					</Tab>
				</Tabs>
			</Modal.Body>
			{showErr && (
				<Modal.Body>
					<Alert variant="danger">{errMsg}</Alert>
				</Modal.Body>
			)}
			<Modal.Footer>
				<div style={{ display: "flex" }}>
					<Button
						size="lg"
						onClick={() => {
							uploadReview()
						}}>
						{" "}
						Upload{" "}
					</Button>
					<h5>
						<Form.Switch
							style={{ marginTop: "8px", marginLeft: "15px" }}
							label="Approve"
							size="lg"
							onChange={() =>
								setApproval((approval) => !approval)
							}
						/>
					</h5>
				</div>
			</Modal.Footer>
		</Modal>
	)
}
