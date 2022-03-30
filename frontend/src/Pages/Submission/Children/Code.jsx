/**
 * Code.jsx
 * Author: 190019931
 *
 * React component for displaying code
 */

import React, { useState, useEffect, useRef } from "react"
import axiosInstance from "../../../Web/axiosInstance"
import {
	Form,
	InputGroup,
	Card,
	Button,
	Collapse,
	Tabs,
	Tab
} from "react-bootstrap"
import { SwitchTransition, CSSTransition } from "react-transition-group"
import MonacoEditor, { monaco } from "react-monaco-editor"
import FadeInTransition from "../../../Components/Transitions/FadeIn.module.css"
import Comments from "./Comments"
import ReactMarkdown from "react-markdown"

const fileEndpoint = "/file"
const defaultLanguage = "javascript"
const defaultTheme = "vs"
const defaultLine = 1

const CODE_HEIGHT = "50vh"

function Code({ id, show }) {
	//File to be fetched
	const [file, setFile] = useState({
		ID: null,
		submissionId: null,
		path: "",
		CreatedAt: "",
		UpdatedAt: ""
	})

	//Code displayed
	const [code, setCode] = useState("")

	//Monaco's states
	const monacoRef = useRef(null)
	const [theme, setTheme] = useState(defaultTheme)
	const [language, setLanguage] = useState(defaultLanguage)
	const [startLine, setStartLine] = useState(defaultLine)
	const [endLine, setEndLine] = useState(defaultLine)
	const [decorations, setDecorations] = useState([])

	//File's comments
	const [comments, setComments] = useState([])
	const [showComments, setShowComments] = useState(false)

	//Fetch file (content and comments) by file ID
	useEffect(() => {
		getFile()
	}, [id]) //If the file ID is changed, new file is fetched and components re render

	//Get the file as specified by the id prop passed
	const getFile = () => {
		if (id && id != -1) {
			axiosInstance
				.get(fileEndpoint + "/" + id)
				.then((response) => {
					//Set file, code and comments
					setFile(response.data.file)
					setCode(atob(response.data.file.base64Value))
					monacoRef?.current?.editor?.setSelection(
						new monaco.Selection(0, 0, 0, 0)
					) //Fixes line issue
					monacoRef?.current?.editor?.layout() //Fixes Code height issue
					setComments(response.data.file.comments)
					getDecorations(response.data.file.comments)
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

	//Gets the comment decorations for the monaco instance
	const getDecorations = (comments) => {
		let newDecorations = comments?.map((comment) => {
			return {
				range: new monaco.Range(
					comment.startLine,
					1,
					comment.endLine,
					1
				),
				options: {
					isWholeLine: true,
					className: "myContentClass",
					glyphMarginClassName: "myGlyphMarginClass",
					hoverMessage: [],
					glyphMarginHoverMessage: [
						{
							value: atob(comment.base64Value)
						}
					]
				}
			}
		})
		setDecorations(
			monacoRef.current
				? monacoRef.current.editor.deltaDecorations(
						decorations,
						newDecorations ? newDecorations : []
				  )
				: []
		)
	}

	//Invoked when the monaco component is mounted, add the in-line commenting functionality here
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

	//Options for the Monaco editor component
	const options = {
		selectOnLineNumbers: true,
		glyphMargin: true,
		readOnly: true
	}

	//
	const codeHTML = () => {
		return (
			<div>
				<div style={{ display: "flex" }}>
					<div style={{ flex: "1", marginRight: "5px" }}>
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
								<option>c</option>
								<option>clojure</option>
								<option>cpp</option>
								<option>csharp</option>
								<option>css</option>
								<option>dart</option>
								<option>dockerfile</option>
								<option>elixir</option>
								<option>fsharp</option>
								<option>go</option>
								<option>graphql</option>
								<option>html</option>
								<option>ini</option>
								<option>java</option>
								<option>javascript</option>
								<option>json</option>
								<option>julia</option>
								<option>kotlin</option>
								<option>lua</option>
								<option>markdown</option>
								<option>mips</option>
								<option>mysql</option>
								<option>objective-c</option>
								<option>pascal</option>
								<option>perl</option>
								<option>pgsql</option>
								<option>php</option>
								<option>plaintext</option>
								<option>powerquery</option>
								<option>powershell</option>
								<option>pug</option>
								<option>python</option>
								<option>qsharp</option>
								<option>r</option>
								<option>razor</option>
								<option>redis</option>
								<option>ruby</option>
								<option>rust</option>
								<option>scala</option>
								<option>scheme</option>
								<option>scss</option>
								<option>shell</option>
								<option>sql</option>
								<option>swift</option>
								<option>typescript</option>
								<option>vb</option>
								<option>xml</option>
								<option>yaml</option>
							</Form.Select>
						</InputGroup>
					</div>
					<div style={{ flex: "1", marginLeft: "5px" }}>
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
								<option value="vs-dark">
									Visual Studio Dark
								</option>
								<option value="hc-black">
									High Contrast Dark
								</option>
							</Form.Select>
						</InputGroup>
					</div>
				</div>
				<MonacoEditor
					ref={monacoRef}
					height={CODE_HEIGHT}
					language={language}
					theme={theme}
					value={code}
					options={options}
					editorDidMount={editorDidMount}
				/>
			</div>
		)
	}

	const pdfHTML = () => {
		return (
			<div style={{ textAlign: "center" }}>
				<embed
					height="700"
					width="80%"
					src={"data:application/pdf;base64," + file.base64Value}
				/>
			</div>
		)
	}

	const imgHTML = (ext) => {
		return (
			<div style={{ textAlign: "center" }}>
				<img
					className="image"
					src={"data:mage/" + ext + ";base64," + file.base64Value}
				/>
			</div>
		)
	}

	const mdHTML = () => {
		return <ReactMarkdown>{atob(file.base64Value)}</ReactMarkdown>
	}

	const renderFile = () => {
		const extension = file.path.split(".").pop()
		switch (extension) {
			case "md":
				return (
					<Tabs defaultActiveKey="code" style={{ margin: "15px" }}>
						<Tab eventKey="code" title="Code">
							{codeHTML()}
						</Tab>
						<Tab eventKey="md" title="Markdown">
							{mdHTML()}
						</Tab>
					</Tabs>
				)
			case "pdf":
				return pdfHTML()
			case "png":
			case "jpg":
			case "jpeg":
				return imgHTML(extension)
			default:
				return codeHTML()
		}
	}

	return (
		<Card>
			<Card.Body>
				<SwitchTransition>
					<CSSTransition
						key={show}
						timeout={100}
						classNames={{ ...FadeInTransition }}>
						{show ? (
							<div
								style={{
									display: "flex",
									justifyContent: "space-between"
								}}>
								<h4>{file.path}</h4>
								<Button
									variant="dark"
									onClick={
										monacoRef.current
											? monacoRef.current.editor._actions
													.Comment._run
											: () => {
													setShowComments(true)
													setStartLine(defaultLine)
													setEndLine(defaultLine)
											  }
									}>
									Show comments
								</Button>
							</div>
						) : (
							<h4>File Viewer</h4>
						)}
					</CSSTransition>
				</SwitchTransition>
				<Collapse in={show}>
					<div>
						{id && id != -1 ? (
							<div>
								<div
									style={{
										display: "flex",
										justifyContent: "space-between"
									}}>
									<Card.Text>
										Created:
										{"\t" +
											new Date(
												file.CreatedAt
											).toDateString()}
										<br />
										Updated:
										{"\t" +
											new Date(
												file.UpdatedAt
											).toDateString()}
									</Card.Text>
									<div className="text-muted">
										Press Ctrl + m to comment on selected
										line
									</div>
								</div>
								{renderFile()}
								<br />
							</div>
						) : (
							<div>No file selected.</div>
						)}
					</div>
				</Collapse>
			</Card.Body>
			<Comments
				id={id}
				comments={comments}
				startLine={startLine}
				endLine={endLine}
				show={showComments}
				setShow={setShowComments}
				refresh={getFile}
			/>
		</Card>
	)
}

export default Code
