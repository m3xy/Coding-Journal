import React, { useState, useEffect } from "react"
import { Card, Button, Container, Alert, Form, Row, Col } from "react-bootstrap"
import { useNavigate } from "react-router-dom"
import { CSSTransition } from "react-transition-group"
import axiosInstance from "../../Web/axiosInstance"
import FormText from "../FormComponents/FormText"
import JwtService from "../../Web/jwt.service"

const defaultMessages = {
	password: "A valid password is required",
	repeatPassword: "Passwords do not match"
}

export default ({ updateUser }) => {
	const [optionals, setOptionals] = useState({
		firstName: "",
		lastName: "",
		organization: ""
	})
	const [form, setForm] = useState({
		password: "",
		repeatPassword: ""
	})
	const navigate = useNavigate()
	const [errors, setErrors] = useState({})
	const [error, setError] = useState(null)
	const [moddedFields, setModdedFields] = useState([])

	useEffect(() => {
		let userID = JwtService.getUserID()
		if (!userID) {
			navigate("/login")
		}
	}, [])

	// Handle form data change
	const handleChange = (e) => {
		// Add the value to the form.
		setForm((form) => {
			return { ...form, [e.target.name]: e.target.value }
		})

		// Add the field to the modifed fields if not done yet.
		if (!moddedFields.includes(e.target.name)) {
			setModdedFields((fields) => {
				return [...fields, e.target.name]
			})
		}

		// Validate field's errors
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
					[e.target.name]: defaultMessages[e.target.name]
				}
			})
		}
	}

	// Show an error alert
	const showError = (err) => {
		setMsg(
			<>
				<b>Edit Failed</b> - {err}
			</>
		)
		setShowAlert(true)
	}

	const handleOptional = (e) => {
		setOptionals((optionals) => {
			return { ...optionals, [e.target.name]: e.target.value }
		})
		return handleChange(e)
	}

	const updateProfile = () => {
		let data = {
			firstName: optionals.changeFirstName,
			lastName: optionals.changeLastName,
			phoneNumber: optionals.changePhoneNum,
			organization: optionals.changeOrg,
			password: form.changePassword
		}

		let submit = (data) => {
			axiosInstance
				.post("/user/" + JwtService.getUserID() + "/edit", data)
				.catch((error) => {
					console.log(error)
					if (error.hasOwnProperty("response")) {
						showError(
							error.response?.data?.message +
								" - " +
								error.response?.status
						)
					} else {
						showError("Please try again later")
					}
				})
		}

		submit(data)
		updateUser()
	}

	const validate = (target, value) => {
		switch (target) {
			case "changeLastName":
				return true
			case "changeFirstName":
				return true
			case "changeOrg":
				return true
			case "changePhoneNum":
				return (
					value.length > 0 &&
					String(value).match(
						/^[\+]?[(]?[0-9]{3}[)]?[-\s\.]?[0-9]{3}[-\s\.]?[0-9]{4,6}$/
					)
				)
			case "changePassword":
				return (
					value.length > 0 &&
					String(value).match(
						/^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*?[#?!@$ %^&*-]).{8,}$/
					) // From https://ihateregex.io/expr/password
				)
			case "repeatPassword":
				return value === form.password
		}
	}

	const deleteProfile = () => {
		axiosInstance
			.post("/user/" + JwtService.getUserID() + "/delete")
			.then((response) => {
				JwtService.rmUser()
				navigate("/login")
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const createAnalyticsAndSettings = (id) => {
		const [showButton, setShowButton] = useState(true)
		const [showMessage, setShowMessage] = useState(false)
		const [key, setKey] = useState("edit")
		return (
			<Container style={{ paddingTop: "2rem" }}>
				<Card style={{ width: "flex" }}>
					<Card.Body>
						<Form>
							<Row>
								<FormText
									display="First Name"
									name="changeFirstName"
									onChange={handleOptional}
									isInvalid={errors.hasOwnProperty(
										"changeFirstName"
									)}
								/>
								<FormText
									display="Last Name"
									name="changeLastName"
									onChange={handleOptional}
									isInvalid={errors.hasOwnProperty(
										"changeLastName"
									)}
								/>
							</Row>
							<Row>
								<FormText
									display="Organization"
									name="changeOrg"
									onChange={handleOptional}
									isInvalid={errors.hasOwnProperty(
										"changeOrg"
									)}
								/>
								<FormText
									display="Phone Number"
									name="changePhoneNum"
									onChange={handleOptional}
									isInvalid={errors.hasOwnProperty(
										"changePhoneNum"
									)}
								/>
							</Row>
							<Row>
								<Col>
									<Button
										variant="danger"
										onClick={() => {
											deleteProfile(
												JwtService.getUserID()
											)
										}}
										style={{ margin: "8px" }}>
										Delete Account
									</Button>
								</Col>{" "}
								<Col>
									<Button
										variant="primary"
										onClick={updateProfile}
										style={{ margin: "8px" }}>
										Save Changes
									</Button>{" "}
								</Col>{" "}
							</Row>
						</Form>
					</Card.Body>
				</Card>

				<CSSTransition
					in={showMessage}
					timeout={300}
					classNames="alert"
					unmountOnExit
					onEnter={() => setShowButton(false)}
					onExited={() => setShowButton(true)}>
					<Alert
						variant="primary"
						dismissible
						onClose={() => setShowMessage(false)}>
						<Alert.Heading>Animated alert message</Alert.Heading>
						<p>
							This alert message is being transitioned in and out
							of the DOM.
						</p>
						<Button onClick={() => setShowMessage(false)}>
							Close
						</Button>
					</Alert>
				</CSSTransition>
			</Container>
		)
	}

	return (
		<div>
			<div>
				{error
					? "Something went wrong, please try again later..."
					: createAnalyticsAndSettings(0)}
			</div>
		</div>
	)
}
