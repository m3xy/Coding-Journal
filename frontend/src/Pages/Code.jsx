/**
 * Code.jsx
 * Author: 190019931
 *
 * React component for displaying code
 */

import React, { useState, useEffect, useRef } from "react"
import axiosInstance from "../Web/axiosInstance"
import { Form, InputGroup, Card, Button } from "react-bootstrap"
import MonacoEditor from "react-monaco-editor"
import Comments from "./Comments"

const fileEndpoint = "/file"
const defaultLanguage = "javascript"
const defaultTheme = "vs"
const defaultLine = 1

function Code({ id }) {
	const [file, setFile] = useState({
		ID: null,
		submissionId: null,
		path: "",
		CreatedAt: "",
		UpdatedAt: ""
	})
	const [code, setCode] = useState("")

	const monacoRef = useRef(null)
	const [theme, setTheme] = useState(defaultTheme)
	const [language, setLanguage] = useState(defaultLanguage)
	const [startLine, setStartLine] = useState(defaultLine)
	const [endLine, setEndLine] = useState(defaultLine)

	const [comments, setComments] = useState([])
	const [showComments, setShowComments] = useState(false)

	useEffect(() => {
		getFile()
	}, [id])

	const getFile = () => {
		if (id && id != -1) {
			axiosInstance
				.get(fileEndpoint + "/" + id)
				.then((response) => {
					//Set file, code and comments
					setFile(response.data.file)
					setCode(atob(response.data.file.base64Value))
					// monacoRef.current.editor.setValue(atob(response.data.file.base64Value)); //Fixes problem with selecting
					setComments(response.data.file.comments)
				})
				.catch((error) => {
					console.log(error)
				})
		} else {
			setFile({
				ID: null,
				submissionId: null,
				path: "",
				CreatedAt: "",
				UpdatedAt: ""
			})
			setComments([])
			setCode("")
		}
	}

	const editorDidMount = (editor, monaco) => {
		editor.addAction({
			id: "Comment", // An unique identifier of the contributed action.
			label: "Comment", // A label of the action that will be presented to the user. (Right-click)
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyM], // An optional array of keybindings for the action.
			precondition: null, // A precondition for this action.
			keybindingContext: null, // A rule to evaluate on top of the precondition in order to dispatch the keybindings.
			contextMenuGroupId: "navigation",
			contextMenuOrder: 1.5,

			// Method that will be executed when the action is triggered.
			// @param editor The editor instance is passed in
			run: function (ed) {
				setStartLine(ed.getSelection().startLineNumber)
				setEndLine(ed.getSelection().endLineNumber)
				setShowComments(true)
			}
		})
		editor.focus()
	}

	const options = {
		selectOnLineNumbers: true,
		glyphMargin: true,
		readOnly: true
	}

	const codeHTML = () => {
		return (
			<>
				<div className="w-50">
					<InputGroup size="sm" className="mb-3">
						<InputGroup.Text id="inputGroup-sizing-sm">
							Language:{" "}
						</InputGroup.Text>
						<Form.Select
							defaultValue={language}
							size="sm"
							onChange={(e) => {
								setLanguage(e.target.value)
							}}>
							<option value="javascript">Javascript</option>
							<option value="html">HTML</option>
							<option value="css">CSS</option>
							<option value="json">JSON</option>
							<option value="java">Java</option>
							<option value="python">Python</option>
						</Form.Select>
					</InputGroup>
				</div>
				<div className="w-50">
					<InputGroup size="sm" className="mb-3">
						<InputGroup.Text id="inputGroup-sizing-sm">
							Theme:{" "}
						</InputGroup.Text>
						<Form.Select
							size="sm"
							onChange={(e) => {
								setTheme(e.target.value)
							}}>
							<option value="vs">Visual Studio</option>
							<option value="vs-dark">Visual Studio Dark</option>
							<option value="hc-black">High Contrast Dark</option>
						</Form.Select>
					</InputGroup>
				</div>
				<div className="w-100">
					<MonacoEditor
						ref={monacoRef}
						height="1000"
						language={language}
						theme={theme}
						value={code}
						options={options}
						editorDidMount={editorDidMount}
					/>
				</div>
			</>
		)
	}

	const pdfHTML = () => {
		return (
			<embed
				height="1000"
				width="100%"
				src={"data:application/pdf;base64," + file.base64Value}
			/>
		)
	}

	return (
		<Card border="light" className="row no-gutters">
			<Card.Header>
				<b>Code</b>
			</Card.Header>
			{id && id != -1 ? (
				<Card.Body>
					<Card.Title>{file.path}</Card.Title>
					<Card.Text>
						Created: {new Date(file.CreatedAt).toDateString()}
					</Card.Text>
					{file.path.split(".").pop() !== "pdf"
						? codeHTML()
						: pdfHTML()}
					<br />
					<Button
						variant="dark"
						onClick={
							monacoRef.current
								? monacoRef.current.editor._actions.Comment._run
								: () => {
										setShowComments(true)
										setStartLine(defaultLine)
										setEndLine(defaultLine)
								  }
						}>
						Show comments
					</Button>
					<Comments
						id={id}
						comments={comments}
                        setComments={setComments}
						startLine={startLine}
						endLine={endLine}
						show={showComments}
						setShow={setShowComments}
						refresh={getFile}></Comments>
				</Card.Body>
			) : (
				<>No file selected</>
			)}
			<Card.Footer className="text-muted">
				Last updated: {new Date(file.UpdatedAt).toDateString()}
			</Card.Footer>
		</Card>
	)
}

export default Code
