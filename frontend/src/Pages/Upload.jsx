/**
 * Upload.jsx
 * author: 190019931
 *
 * Page for uploading files
 */

import React, { useState, useEffect } from "react"
import DragAndDrop from "./DragAndDrop"
import axiosInstance from "../Web/axiosInstance"
import {
	Form,
	Button,
	Card,
	ListGroup,
	CloseButton,
	FloatingLabel,
	Container,
	Row,
	Col,
	InputGroup,
	FormControl,
	Tabs,
	Tab
} from "react-bootstrap"
import JwtService from "../Web/jwt.service"
import { useNavigate } from "react-router-dom"

const uploadEndpoint = "/submissions/create"

class Upload extends React.Component {
	constructor(props) {
		super(props)

		this.state = {
			authors: [], //Change to string
			files: [],
			submissionName: "",
			submissionAbstract: "",
			tags: []
		}

		this.dropFiles = this.dropFiles.bind(this)
		this.handleSubmit = this.handleSubmit.bind(this)
		this.handleDrop = this.handleDrop.bind(this)
		this.setSubmissionName = this.setSubmissionName.bind(this)
		this.setSubmissionAbstract = this.setSubmissionAbstract.bind(this)
		this.tagsInput = React.createRef()
	}

	componentDidMount() {
		console.log(JwtService.getUserID())
	}

	dropFiles(e) {
		this.handleDrop(e.target.files)
	}

	setSubmissionName(e) {
		this.setState({
			submissionName: e.target.value
		})
	}

	setSubmissionAbstract(e) {
		this.setState({
			submissionAbstract: e.target.value
		})
	}

	/**
	 * Sends a POST request to the go server to upload a submission
	 *
	 * @param {JSON} authors Submission files' Authors' User ID
	 * @param {Array.<File>} files Submission files
	 * @returns
	 */
	uploadSubmission(
		authors,
		submissionName,
		submissionAbstract,
		files,
		categories
	) {
		console.log(authors)
		// const authorID = JSON.parse(userId).userId;  //Extract author's userId

		const filePromises = files.map((file, i) => {
			//Create Promise for each file (Encode to base 64 before upload)
			return new Promise((resolve, reject) => {
				files[i].path = files[i].name
				const reader = new FileReader()
				reader.readAsDataURL(file)
				reader.onload = function (e) {
					files[i].base64Value = e.target.result.split(",")[1]
					resolve() //Promise(s) resolved/fulfilled once reader has encoded file(s) into base 64
				}
				reader.onerror = function () {
					reject()
				}
			})
		})

		Promise.all(filePromises).then(() => {
			let data = {
				name: submissionName,
				license: "MIT",
				abstract: submissionAbstract,
				files: files,
				authors: authors,
				categories: categories
			}
			console.log(data)
			axiosInstance
				.post(uploadEndpoint, data)
				.then((response) => {
					console.log(response)
					console.log("Submission ID: " + response.data["ID"])

					var submissionPage = window.open(
						"/submission/" + response.data["ID"]
					)
				})
				.catch((error) => {
					console.log(error)
				})
		})
	}

	handleSubmit(e) {
		e.preventDefault()

		//Checking there are files to submit
		if (this.state.files.length === 0) {
			return
		}

		let userId = JwtService.getUserID() //Preparing to get userId

		if (userId === null) {
			//If user has not logged in, disallow submit
			console.log("Not logged in")
			return
		}

		// this.state.submissionName = this.state.files[0].name;     //Temp, 1 file uploads
		// console.log(this.state.submissionName);

		// this.props.upload(userId, this.state.submissionName, this.state.files).then((submission) => {
		//    console.log("Submission ID: " + submission);
		//    var codePage = window.open("/code");
		//    codePage.submission = submission;
		//    codePage.submissionName = this.state.submissionName;
		// }, (error) => {
		//    console.log(error);
		// });

		this.state.authors.push(userId)
		this.uploadSubmission(
			this.state.authors,
			this.state.submissionName,
			this.state.submissionAbstract,
			this.state.files,
			this.state.tags
		)

		//Clearing form
		document.getElementById("formFile").files = new DataTransfer().files
		document.getElementById("submissionName").value = ""
		this.setState({
			authors: [],
			files: [],
			submissionName: "",
			submissionAbstract: "",
			tags: []
		})

		console.log("Files submitted")
	}

	handleDrop(files) {
		// console.log(files);
		// console.log(this.state.files);
		// console.log(document.getElementById("formFile").files);

		if (this.state.files.length === 1)
			return /* Remove later for multiple files */

		// if(this.writer.userId === null){
		//     console.log("Not logged in!");
		//     return;
		// }

		let formFileList = new DataTransfer()
		let fileList = this.state.files

		for (let i = 0; i < files.length; i++) {
			if (
				!files[i]
				// || !(files[i].name.endsWith(".css") || files[i].name.endsWith(".java") || files[i].name.endsWith(".js"))
			) {
				console.log("Invalid file")
				return
			}

			for (let j = 0; j < fileList.length; j++) {
				if (files[i].name === fileList[j].name) {
					console.log("Duplicate file")
					return
				}
			}

			console.log(files[i])
			fileList.push(files[i])
			formFileList.items.add(files[i])
		}

		document.getElementById("formFile").files = formFileList.files
		this.setState({ files: fileList })
	}

	removeFile(key) {
		let formFileList = new DataTransfer()
		let fileList = this.state.files

		for (let i = 0; i < this.state.files.length; i++) {
			formFileList.items.add(this.state.files[i])
		}
		formFileList.items.remove(key)
		fileList.splice(key, 1)

		document.getElementById("formFile").files = formFileList.files
		this.setState({
			files: fileList
		})
	}

	render() {
		const files = this.state.files.map((file, i) => {
			return (
				<ListGroup.Item as="li" key={i} action onClick={() => {}}>
					<CloseButton
						onClick={() => {
							this.removeFile(i)
						}}
					/>
					<br />
					<label>File name: {file.name}</label>
					<br />
					<label>File type: {file.type}</label>
					<br />
					<label>File Size: {file.size} bytes</label>
					<br />
					<label>
						Last modified:{" "}
						{new Date(file.lastModified).toUTCString()}
					</label>
				</ListGroup.Item>
			)
		})

		const tags = this.state.tags.map((tag, i) => {
			return (
				<Button
					key={i}
					variant="outline-secondary"
					size="sm"
					onClick={() => {
						this.setState({
							tags: this.state.tags.filter(
								(value) => value !== tag
							)
						})
					}}>
					{tag}
				</Button>
			)
		})

		return (
			<Container>
				<br />
				<Row>
					<Col></Col>
					<Col xs={4}>
						<DragAndDrop handleDrop={this.handleDrop}>
							<Card>
								<Form onSubmit={this.handleSubmit}>
									<Card.Header className="text-center">
										<h5>Submission Upload</h5>
									</Card.Header>
									<Tabs
										justify
										defaultActiveKey="details"
										id="profileTabs"
										className="mb-3">
										<Tab eventKey="details" title="Details">
											<Row>
												<Form.Group
													className="mb-3"
													controlId="submissionName">
													<Form.Label>
														Name
													</Form.Label>
													<Form.Control
														type="text"
														placeholder="Name"
														required
														onChange={
															this
																.setSubmissionName
														}
													/>
												</Form.Group>
												<Form.Group
													className="mb-3"
													controlId="submissionAbstract">
													<Form.Label>
														Abstract
													</Form.Label>
													<Form.Control
														as="textarea"
														rows={3}
														placeholder="Abstract"
														required
														onChange={
															this
																.setSubmissionAbstract
														}
													/>
												</Form.Group>
											</Row>
											<Row>
												<InputGroup className="mb-3">
													<FormControl
														placeholder="Add tags here"
														aria-label="Tags"
														aria-describedby="addTag"
														ref={this.tagsInput}
													/>
													<Button
														variant="outline-secondary"
														id="addTag"
														onClick={() => {
															if (
																this.state.tags.includes(
																	this
																		.tagsInput
																		.current
																		.value
																)
															)
																return
															this.setState({
																tags: [
																	...this
																		.state
																		.tags,
																	this
																		.tagsInput
																		.current
																		.value
																]
															})
															this.tagsInput.current.value =
																""
														}}>
														Add
													</Button>
												</InputGroup>
												<Col>{tags}</Col>
											</Row>
										</Tab>
										<Tab eventKey="files" title="Files">
											<Row>
												<Form.Group
													controlId="formFile"
													className="mb-3">
													<Form.Control
														type="file"
														accept=".css,.java,.js"
														required
														onChange={
															this.dropFiles
														}
													/>
													{/* multiple later w/ "zip,application/octet-stream,application/zip,application/x-zip,application/x-zip-compressed" */}
												</Form.Group>
												<Card.Body>
													{this.state.files.length >
													0 ? (
														<ListGroup>
															{files}
														</ListGroup>
													) : (
														<Card.Text
															className="text-center"
															style={{
																color: "grey"
															}}>
															<i>
																Drag and Drop{" "}
																<br />
																here
															</i>
															<br />
															<br />
														</Card.Text>
													)}
												</Card.Body>
											</Row>
										</Tab>
									</Tabs>
									<Card.Footer className="text-center">
										<Button
											variant="outline-secondary"
											type="submit">
											Upload
										</Button>{" "}
									</Card.Footer>
								</Form>
							</Card>
						</DragAndDrop>
					</Col>
					<Col></Col>
				</Row>
			</Container>
		)
	}
}

const UploadFunc = () => {
	const [form, setForm] = useState({
		file: { path: "", base64Value: "" },
		authors: [],
		submissionName: ""
	})
	const [optionals, setOptionals] = useState({
		submissionAbstract: "",
		tags: []
	})
	const [errors, setErrors] = useState({})
	const [moddedFields, setModdedFields] = useState([])

	const tagInput = React.useRef()

	const navigate = useNavigate()

	useEffect(() => {
		console.log(JwtService.getUserID())
	}, [])

	// Validate an element inserted into the form.
	const validate = (key, val) => {
		switch (key) {
			case submissionAbstract:
				return true
			case file:
				return String(val.name).match(
					/^([A-z0-9-_+]+\/)*([A-z0-9-_+]+\.(zip))$/
				)
			case submissionName:
			case tag:
			case author:
				return val.length > 1
		}
	}

	const handleDrop = (e) => {
		let files = e.target.files
		if (files.length > 1) {
			return false
		}
		let file = files[0]
		if (!validate("file", file.name)) {
			return false
		}

		let formFileList = new DataTransfer()
		handleRequired({ target: { name: "file", file } })
	}

	// Handle a change in a form value handled by required and optional.
	const handleChange = (e) => {
		const validated = validate(e.target.name, e.target.value)
		if (validated && errors.hasOwnProperty(e.target.name)) {
			setErrors((errors) => {
				const newErrs = { ...errors }
				delete newErrs[e.target.name]
				return newErrs
			})
		} else if (!validated && !errors.hasOwnProperty(e.target.name)) {
			setErrors((errors) => {
				return { ...errors, [e.target.name]: true }
			})
		}
	}

	// Handler for required form content.
	const handleRequired = (e) => {
		setForm((form) => {
			return { ...form, [e.target.name]: e.target.value }
		})

		if (!moddedFields.includes(e.target.name)) {
			setModdedFields((fields) => {
				return [...fields, e.target.name]
			})
		}
		return handleChange(e)
	}

	// Handler for optional form content (e.g. abstract, tags).
	const handleOptional = (e) => {
		setOptionals((optionals) => {
			return { ...optionals, [e.target.name]: e.target.value }
		})
		return handleChange(e)
	}

	// Handler for the submit button - submit form, upload submission.
	const handleSubmit = (e) => {
		e.preventDefault()
		uploadSubmission(
			form.authors,
			form.submissionName,
			optionals.submissionAbstract,
			form.file,
			optionals.tags
		)
	}

	const uploadSubmission = async (authors, name, subAbstract, file, tags) => {
		// Get files
		try {
			let data = {
				name: name,
				authors: authors,
				abstract: subAbstract,
				tags: tags,
				base64Value: ""
			}
			// Get zip's encoded value
			await new Promise((resolve, reject) => {
				file.path = file.name
				const reader = new FileReader()
				reader.readAsDataURL(file)
				reader.onload = (e) => {
					data.base64Value = e.target.result.split(",")[1]
					resolve()
				}
				reader.onerror = () => {
					reject(reader.error)
				}
			})
		} catch (err) {
			console.log(err)
			return
		}

		// Send axios request
		axiosInstance
			.post(uploadEndpoint, data)
			.then((response) => {
				navigate("/submission/" + response.data.ID)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const removeFile = () => {
		setForm((form) => {
			return { ...form, file: "" }
		})
	}

	const addTag = (e) => {
		tagInput = e.target.value
	}

	const submitTag = (e) => {
		setOptionals((optionals) => {
			let newOpt = optionals
			delete optionals.tags
			return { ...newOpt, tags: [...optionals.tags, e.target.value] }
		})
		setTagInput("")
	}

    // Get a tag button per given tag in optionals.
	let tags = optionals.tags.map((tag, i) => {
		return (
			<Button
				key={i}
				variant="outline-secondary"
				size="sm"
				onClick={() => {
					setForm((form) => {
						let newForm = form
						newForm.tags = form.tags.filter((value) => {
							value !== tag
						})
						return newForm
					})
				}}>
				{tag}
			</Button>
		)
	})

    // The card corresponding to a file.
	let fileCard = (
		<ListGroup.Item as="li" action onClick={() => {}}>
			<CloseButton onClick={removeFile} />
			<br />
			<label>File name: {form.file.name}</label>
			<br />
			<label>File type: {form.file.type}</label>
			<br />
			<label>File Size: {form.file.size} bytes</label>
			<br />
			<label>
				Last modified: {new Date(form.file.lastModified).toUTCString()}
			</label>
		</ListGroup.Item>
	)

    // The tab with the submission's details form.
	let detailsTab = (
		<Tab eventKey="details" title="details">
			<Row>
				<Form.Group className="mb-3" controlId="submissionName">
					<Form.Label>Name</Form.Label>
					<Form.Control
						type="text"
						name="submissionName"
						isInvalid={errors.hasOwnProperty("submissionName")}
						onChange={handleRequired}
					/>
				</Form.Group>

				<Form.Group className="mb-3" controlId="submissionAbstract">
					<Form.Label>Abstract</Form.Label>
					<Form.Control
						as="textarea"
						rows={5}
						placeholder="Write/Copy abstract here."
						isInvalid={errors.hasOwnProperty("submissionAbstract")}
						onChange={handleOptional}
					/>
				</Form.Group>
			</Row>
			<Row>
				<InputGroup>
					<FormControl
						placeholder="Add tags here"
						aria-label="Tags"
						aria-describedby="addTags"
						ref={tagInput}
					/>
					<Button
						variant="outline-secondary"
						id="addTag"
						onClick={submitTag}
					/>
				</InputGroup>
				<Col>{tags}</Col>
			</Row>
		</Tab>
	)

    // The tab for file selection.
	let filesTab = (
		<Tab eventKey="files" title="Files">
			<Row>
				<Form.Group controlId="formFile" className="mb-3">
					<Form.Control
						type="file"
						accept=".zip"
						required
						onChange={handleDrop}
					/>
					<Card.Body>
						{form.files != "" ? (
							<ListGroup>{fileCard}</ListGroup>
						) : (
							<Card.Text
								className="text-center"
								style={{ color: "grey" }}>
								<i>
									Drag and Drop <br />
								</i>
								<br />
								<br />
							</Card.Text>
						)}
					</Card.Body>
				</Form.Group>
			</Row>
		</Tab>
	)

	let uploadCard = (
		<Card>
			<Form onSubmit={handleSubmit}>
				<Card.Header>Submission Upload</Card.Header>
				<Tabs
					justify
					defaultActiveKey="details"
					id="profileTabs"
					className="mb-3">
					{detailsTab}
					{filesTab}
				</Tabs>
				<Card.Footer>
					<Button
						variant="outline-secondary"
						type="submit"
						disabled={
							Object.keys(errors).length === 0 &&
							moddedFields.length === Object.keys(form).length
						}>
						Upload
					</Button>{" "}
				</Card.Footer>
			</Form>
		</Card>
	)
    
    return (
        <Container>
            <br />
            <Row>
                <Col></Col>
                <Col xs={4}>
                    <DragAndDrop handleDrop={handleDrop}>
                        {uploadCard}
                    </DragAndDrop>
                </Col>
                <Col></Col>
            </Row>
        </Container>
    )
}

export default UploadFunc
