import React, { useState, useEffect } from "react"
import axiosInstance from "../../Web/axiosInstance"
import { Form, Button, Card, Tabs, Tab, Alert } from "react-bootstrap"
import JwtService from "../../Web/jwt.service"
import { FormText, FormAdder, FormFile } from "../../Components/FormComponents"
import { useNavigate } from "react-router-dom"
import { CSSTransition } from "react-transition-group"
import styles from "./Upload.module.css"
import FadeInStyles from "../../Components/Transitions/FadeIn.module.css"

const defaultMsgs = {
	submissionName: "A name is required!",
	files: "A single ZIP file is allowed.",
	tag: "This tag has already been inserted into the submission!"
}

const Upload = () => {
	const [form, setForm] = useState({
		files: [],
		authors: [],
		submissionName: ""
	})
	const [optionals, setOptionals] = useState({
		submissionAbstract: "",
		tags: []
	})
	const [errors, setErrors] = useState({})
	const [moddedFields, setModdedFields] = useState([])

	const [show, setShowAlert] = useState(false)
	const [msg, setMsg] = useState()

	const navigate = useNavigate()

	// Immediately exit if the user is not logged in.
	useEffect(() => {
		let user = JwtService.getUserID()
		if (!user) {
			navigate("/login")
		} else {
			setForm((form) => {
				return { ...form, ["authors"]: [user] }
			})
			setModdedFields((fields) => {
				return [...fields, "authors"]
			})
		}
	}, [])

	// Validate an element inserted into the form.
	const validate = (key, val) => {
		switch (key) {
			// Simply return true for optional, requirementless entries.
			case "tags":
			case "submissionAbstract":
				return true
			// Enforce ZIP format for file name.
			case "file":
				return /^([A-z0-9-_+]+\.(zip))$/.test(val.name)
			case "files":
				return !val
					.map((file) => {
						return validate("file", file)
					})
					.includes(false)
			case "submissionName":
				return val.length > 0 && val.length < 127
			case "tag":
			case "author":
				return val.length > 0
			default:
				return false
		}
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
				return {
					...errors,
					[e.target.name]: defaultMsgs.hasOwnProperty(e.target.name)
						? defaultMsgs[e.target.name]
						: "Invalid."
				}
			})
		}
	}

	// Show an error alert
	const showError = (err) => {
		setMsg(
			<>
				<bold>Upload failed</bold> - {err}
			</>
		)
		setShowAlert(true)
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

	// Upload the submission to the database.
	const uploadSubmission = async () => {
		let data = {
			name: form.submissionName,
			authors: form.authors,
			abstract: optionals.submissionAbstract,
			tags: optionals.tags
		}

		// Send axios request
		let submit = (data) => {
			axiosInstance
				.post("/submissions/create", data)
				.then((response) => {
					navigate("/submission/" + response.data.ID)
				})
				.catch((error) => {
					console.log(error)
					if (error.hasOwnProperty("repsonse")) {
						showError(
							error.response.data.message +
								" - " +
								error.response.status
						)
					} else {
						showError("Please try again later - 500")
					}
				})
		}

		const reader = new FileReader()
		reader.readAsDataURL(form.files[0])
		reader.onloadend = () => {
			let base64 = reader.result.replace("data:", "").replace(/^.+,/, "")
			console.log(base64)
			submit({ ...data, base64: base64 })
		}
		reader.onerror = () => {
			console.log(reader.error)
			showError(reader.error.message)
		}
	}

	// The tab with the submission's details form.
	let detailsTab = (
		<Tab eventKey="details" title="details">
			<FormText
				display="Name"
				name="submissionName"
				isInvalid={errors.hasOwnProperty("submissionName")}
				feedback={
					errors.hasOwnProperty("submissionName")
						? errors.submissionName
						: ""
				}
				onChange={handleRequired}
				required
			/>
			<FormText
				display="Abstract"
				name="submissionAbstract"
				rows={3}
				isInvalid={errors.hasOwnProperty("submissionAbstract")}
				feedback={""}
				onChange={handleOptional}
				as="textarea"
			/>
			<FormAdder
				display="Tags"
				elemName="tag"
				arrName="tags"
				label="addTags"
				feedback={defaultMsgs.tag}
				onChange={handleOptional}
				validate={validate}
			/>
		</Tab>
	)

	// The tab for file selection.
	let filesTab = (
		<Tab eventKey="files" title="Files">
			<FormFile
				accept=".zip"
				display="Submission ZIP"
				name="files"
				elemName="file"
				fileLimit={1}
				validate={validate}
				onChange={handleRequired}
			/>
		</Tab>
	)

	return (
		<>
			<Card className={styles.UploadCard}>
				<Form>
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
							onClick={uploadSubmission}
							disabled={
								Object.keys(errors).length !== 0 ||
								moddedFields.length !== Object.keys(form).length
							}>
							Upload
						</Button>{" "}
					</Card.Footer>
				</Form>
			</Card>
			<CSSTransition
				in={show}
				timeout={100}
				unmountOnExit
				classNames={{ ...FadeInStyles }}>
				<div className={styles.UploadAlert}>
					<Alert
						variant="danger"
						onClose={() => setShowAlert(false)}
						dismissible>
						<p>{msg}</p>
					</Alert>
				</div>
			</CSSTransition>
		</>
	)
}

export default Upload
