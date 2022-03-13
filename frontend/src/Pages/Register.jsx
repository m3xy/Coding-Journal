/*
 * Register.jsx
 * Author: 190019931
 *
 * This file stores the info for rendering the Register page of our Journal
 */

import React, { useState } from "react"
import { Form, Button, Container, Row, Col } from "react-bootstrap"
import axiosInstance from "../Web/axiosInstance"
import { useNavigate } from "react-router-dom"
import JwtService from "../Web/jwt.service.js"

const registerEndpoint = "/auth/register"

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
			.catch((error) => {
				console.log(error)
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
				return { ...errors, [e.target.name]: e.target.value }
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
					<br />
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
								A first name is required!
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
								A last name is required!
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
								We'll never share your email with anyone else.
							</Form.Text>
							<Form.Control.Feedback type="invalid">
								This email is invalid!
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
							/>
							<Form.Control.Feedback type="invalid">
								The password must be valid!
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
							/>
							<Form.Control.Feedback type="invalid">
								Passwords don't match!
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
