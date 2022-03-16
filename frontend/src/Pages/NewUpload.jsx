import React, { useState, useEffect } from "react"
import axiosInstance from "../Web/axiosInstance"
import {
	Form,
	Button,
	Card,
	ListGroup,
	CloseButton,
	Container,
	Row,
	Col,
	InputGroup,
	FormControl,
	Tabs,
	Tab
} from "react-bootstrap"
import JwtService from "../Web/jwt.service"
import Dropzone from "react-dropzone"
import { FormText, FormAdder, FormFile } from "../Components/FormComponents"
import { useNavigate } from "react-router-dom"

const Upload = () => {
	const [form, setForm] = useState({
		file: "",
		authors: [],
		submissionName: ""
	})
	const [optionals, setOptionals] = useState({
		submissionAbstract: "",
		tags: []
	})
	const [errors, setErrors] = useState({})
	const [moddedFields, setModdedFields] = useState([])

	const navigate = useNavigate()

	// Validate an element inserted into the form.
	const validate = (key, val) => {
		switch (key) {
			case "submissionAbstract":
				return true
			case "file":
				return String(val.name).match(
					/^([A-z0-9-_+]+\/)*([A-z0-9-_+]+\.(zip))$/
				)
			case "submissionName":
			case "tag":
			case "author":
				return val.length > 0
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
				<FormText
					display="Name"
					name="submissionName"
					isInvalid={errors.hasOwnProperty("submissionName")}
					onChange={handleRequired}
				/>
				<FormText
					display="Abstract"
					name="submissionAbstract"
					rows={3}
					isInvalid={errors.hasOwnProperty("submissionAbstract")}
					onChange={handleOptional}
					as="textarea"
				/>
			</Row>
			<Row>
				<FormAdder
					display="Tags"
					elemName="tag"
					arrName="tags"
					label="addTags"
					setForm={setForm}
					validate={validate}
				/>
			</Row>
		</Tab>
	)

	// The tab for file selection.
	let filesTab = (
		<Tab eventKey="files" title="Files">
			<Row>
				<FormFile 
					accept=".zip"
					display="Submission ZIP"
					name="file"
					elemName="file"
					fileLimit={1}
					validate={validate}
					setForm={setForm}
				/>
			</Row>
		</Tab>
	)

	return (
		<Container>
			<br />
			<Row>
				<Col></Col>
				<Col xs={4}>
					<Card>
						<Form onSubmit={handleSubmit}>
							<Card.Header>Submission Upload</Card.Header>
							<Card.Body>
								<Tabs
									justify
									defaultActiveKey="details"
									id="profileTabs"
									className="mb-3">
									{detailsTab}
									{filesTab}
								</Tabs>
							</Card.Body>
							<Card.Footer className="text-center">
								<Button
									variant="outline-secondary"
									type="submit"
									disabled={
										Object.keys(errors).length === 0 &&
										moddedFields.length ===
											Object.keys(form).length
									}>
									Upload
								</Button>{" "}
							</Card.Footer>
						</Form>
					</Card>
				</Col>
				<Col></Col>
			</Row>
		</Container>
	)
}

export default Upload
