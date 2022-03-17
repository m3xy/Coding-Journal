import React, { useState, useEffect } from "react"
import axiosInstance from "../../Web/axiosInstance"
import { Form, Button, Card, Container, Tabs, Tab } from "react-bootstrap"
import JwtService from "../../Web/jwt.service"
import { FormText, FormAdder, FormFile } from "../../Components/FormComponents"
import { useNavigate } from "react-router-dom"
import styles from "./Upload.module.css"

const defaultMsgs = {
	submissionName: "A name is required!",
	file: "A single ZIP file is allowed.",
	tag: "This tag has already been inserted into the submission!"
}

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
				return String(val.name).match(/^([A-z0-9-_+]+\.(zip))$/)
			case "submissionName":
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
	const handleSubmit = () => {
		uploadSubmission()
	}

	const uploadSubmission = async () => {
		// Get files
		try {
			let data = {
				name: form.submissionName,
				authors: form.authors,
				abstract: optionals.submissionAbstract,
				tags: optionals.tags,
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
				name="file"
				elemName="file"
				fileLimit={1}
				validate={validate}
				onChange={handleRequired}
			/>
		</Tab>
	)

	return (
		<Container className={styles.UploadContainer}>
			<Card className={styles.UploadCard}>
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
								Object.keys(errors).length === 0 ||
								moddedFields.length !== Object.keys(form).length
							}>
							Upload
						</Button>{" "}
					</Card.Footer>
				</Form>
			</Card>
		</Container>
	)
}

export default Upload
