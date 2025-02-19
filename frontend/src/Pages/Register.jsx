/*
 * Register.jsx
 * Author: 190019931
 *
 * This file stores the info for rendering the Register page of our Journal
 */

import React, { useState } from "react"
import { Form, Button, Container, Row, Col, Alert } from "react-bootstrap"
import axiosInstance from "../Web/axiosInstance"
import { useNavigate } from "react-router-dom"
import JwtService from "../Web/jwt.service.js"

const registerEndpoint = "/auth/register"

const defaultMessages = {
	firstName: "A first name is required.",
	lastName: "A last name is required.",
	email: "A valid email is required.",
	password: "A valid password is required",
	repeatPassword: "Passwords do not match"
}
const DEFAULT_500_MSG = "Registration failed - please try again later."

function Register() {
	const navigate = useNavigate()
	const [form, setForm] = useState({
		firstName: "",
		lastName: "",
		email: "",
		password: "",
		repeatPassword: ""
	})
	const [errors, setErrors] = useState({})
	const [moddedFields, setModdedFields] = useState([])

	const [show, setShow] = useState(false)
	const [alertMsg, setAlertMsg] = useState("")

	const onRegisterSuccessful = () => {
		axiosInstance
			.post("/auth/login", { email: email, password: password })
			.then((response) => {
				JwtService.setUser(
					response.data.access_token,
					response.data.refresh_token
				)
				navigate("/")
			})
			.catch((error) => {
				console.log(error)
				navigate("/login")
			})
	}

	const registerUser = (firstName, lastName, email, password) => {
		let data = {
			firstName: firstName,
			lastName: lastName,
			email: email,
			password: password
		}

		// Send register request to backend.
		axiosInstance
			.post(registerEndpoint, data)
			.then(() => {
				onRegisterSuccessful()
			})
			// Handle errors
			.catch((error) => {
				console.log(error)
				let message = ""

				// Erroring if the error has a server response
				if (error.response != null) {
					let data = error.response.data

					// Get wrong form fields if there are any.
					if (data.hasOwnProperty("fields")) {
						for (let i in data.fields) {
							let field = data.fields[i]
							if (Object.keys(form).includes(field.field)) {
								setErrors((errors) => {
									return {
										...errors,
										[field.field]: field.message
									}
								})
							}
						}
					}
					message = data.message
				} else {
					// Handling of no-response
					message = DEFAULT_500_MSG
				}

				setShow(true)
				setAlertMsg(message)
			})
	}

	const validate = (target, value) => {
		switch (target) {
			case "firstName":
			case "lastName":
				return value.length > 0
			case "email":
				if (value.length == 0) {
					return false
				} else {
					return String(value)
						.toLowerCase()
						.match(
							/^(([^<>()\\[\]\\.,;:\s@"]+(\.[^<>()\\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
						) // From https://emailregex.com/
				}
			case "password":
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

	// Handle form data change
	const handleChange = async (e) => {
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

	// Handle registration form submission, making sure all fields are correct.
	const handleSubmit = async (e) => {
		e.preventDefault()
		registerUser(form.firstName, form.lastName, form.email, form.password)
	}

	return (
		<Container>
			<Row>
				<Col></Col>
				<Col xs={4}>
					{show ? (
						<Alert
							variant="danger"
							onClose={() => setShow(false)}
							dismissible>
							<p>{alertMsg}</p>
						</Alert>
					) : (
						<br />
					)}
					<h2>Register</h2>
					<Form onSubmit={handleSubmit}>
						<Form.Group className="mb-3" controlId="firstName">
							<Form.Label>First Name</Form.Label>
							<Form.Control
								type="text"
								name="firstName"
								placeholder="Enter first name"
								onChange={handleChange}
								isInvalid={errors.hasOwnProperty("firstName")}
							/>
							<Form.Control.Feedback type="invalid">
								{errors["firstName"]}
							</Form.Control.Feedback>
						</Form.Group>

						<Form.Group className="mb-3" controlId="lastName">
							<Form.Label>Last Name</Form.Label>
							<Form.Control
								type="text"
								name="lastName"
								placeholder="Enter last name"
								onChange={handleChange}
								isInvalid={errors.hasOwnProperty("lastName")}
							/>
							<Form.Control.Feedback type="invalid">
								{errors["lastName"]}
							</Form.Control.Feedback>
						</Form.Group>

						<Form.Group className="mb-3" controlId="email">
							<Form.Label>Email address</Form.Label>
							<Form.Control
								type="email"
								name="email"
								placeholder="Enter email"
								onChange={handleChange}
								isInvalid={errors.hasOwnProperty("email")}
							/>
							<Form.Text className="text-muted">
								{!errors.hasOwnProperty("email")
									? "We'll never share your email with anyone else."
									: ""}
							</Form.Text>
							<Form.Control.Feedback type="invalid">
								{errors["email"]}
							</Form.Control.Feedback>
						</Form.Group>

						<Form.Group className="mb-3" controlId="password">
							<Form.Label>Password</Form.Label>
							<Form.Control
								type="password"
								name="password"
								placeholder="Password"
								onChange={handleChange}
								isInvalid={errors.hasOwnProperty("password")}
								isValid={
									!errors.hasOwnProperty("password") &&
									moddedFields.includes("password")
								}
							/>
							<Form.Control.Feedback type="invalid">
								{errors["password"]}
							</Form.Control.Feedback>
						</Form.Group>
						<Form.Group className="mb-3" controlId="repeatPassword">
							<Form.Label>Repeat Password</Form.Label>
							<Form.Control
								type="password"
								name="repeatPassword"
								placeholder="Repeat Password"
								onChange={handleChange}
								isInvalid={errors.hasOwnProperty(
									"repeatPassword"
								)}
								isValid={
									!errors.hasOwnProperty("repeatPassword") &&
									!errors.hasOwnProperty("password") &&
									moddedFields.includes("repeatPassword")
								}
							/>
							<Form.Control.Feedback type="invalid">
								{errors["repeatPassword"]}
							</Form.Control.Feedback>
						</Form.Group>

						<Button
							variant="primary"
							type="submit"
							disabled={
								moddedFields.length !==
									Object.keys(form).length ||
								Object.keys(errors).length !== 0
							}>
							Register
						</Button>
					</Form>
				</Col>
				<Col></Col>
			</Row>
		</Container>
	)
}

export default Register
