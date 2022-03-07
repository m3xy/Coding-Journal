/*
 * LoggedOutForm.jsx
 * Author: 190014935
 *
 * Log In and Register buttons form for guest users.
 */

import React from "react"
import { useNavigate } from "react-router-dom"

import {
	Form,
	Button 
} from "react-bootstrap"

const LoggedOutForm = () => {
	const navigate = useNavigate()
	return(
		<Form>
			<Button
				onClick={() => {
					navigate("/register")
				}}
				variant="primary">
				Register
			</Button>{" "}
			<Button
				onClick={() => {
					navigate("/login")
				}}
				variant="outline-primary">
				Log In
			</Button>{" "}
		</Form>
	)
}
export default LoggedOutForm
