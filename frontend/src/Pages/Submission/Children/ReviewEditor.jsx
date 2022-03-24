import React, { useEffect, useState, useRef } from "react"
import { Editor } from "react-draft-wysiwyg"
import { stateToMarkdown } from "draft-js-export-markdown"
import { EditorState } from "draft-js"
import editorStyles from "./Submission.module.css"
import "react-draft-wysiwyg/dist/react-draft-wysiwyg.css"
import { Modal, Form, Alert, Button } from "react-bootstrap"
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
		setMarkdown(stateToMarkdown(editorState.getCurrentContent()))
	}, [editorState])

	const uploadReview = () => {
		console.log(markdown)
		axiosInstance
			.post("/submission/" + id + "/review", {
				approved: approve,
				base64Value: window.btoa(markdown)
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

	useEffect(() => { }, [editorState])

	return (
		<Modal size="lg" show={show} onHide={() => setShow(false)}>
			<Modal.Header closeButton>
				<h3>
					<div>Review</div>
				</h3>
			</Modal.Header>
			<Modal.Body>
				<Form.Switch
					label="Approve"
					size="lg"
					onChange={(e) =>
						setApproval(e.target.value === "on" ? true : false)
					}
				/>
				<div className={editorStyles.editor}>
					<Editor
						editorState={editorState}
						toolbarClassName="toolbarClassName"
						wrapperClassname="wrapperClassName"
						editorClassName="editorClassName"
						onEditorStateChange={setEditorState}
					/>
				</div>
				<Button
					onClick={() => {
						uploadReview()
					}}>
					{" "}
					Upload{" "}
				</Button>
			</Modal.Body>
			{showErr && (
				<Modal.Body>
					<Alert variant="danger">{errMsg}</Alert>
				</Modal.Body>
			)}
		</Modal>
	)
}
