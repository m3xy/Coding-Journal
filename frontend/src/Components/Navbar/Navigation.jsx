/**
 * Navigation.jsx
 * Author: 190010714, 190019931
 *
 * Navigation Bar Component Class
 */
import React, { useState, useEffect } from "react"
import {
	Navbar,
	Nav,
	Container,
	Form,
	Button,
	Modal
} from "react-bootstrap"
import { useNavigate } from "react-router-dom"
import { LinkContainer } from "react-router-bootstrap"

// User imports
import HeadOffCanvas from "./HeadOffCanvas"
import LoggedInDropdown from "./LoggedInDropdown"
import LoggedOutForm from "./LoggedOutForm"
import { LSearchBar, XSSearchBar } from "./SearchBar"
import JwtService from "../../Web/jwt.service"
import axiosInstance from "../../Web/axiosInstance"

function Navigation() {
	const [user, setUser] = useState(null)
	const [loading, setLoading] = useState(true)
	const navigate = useNavigate()

	// Check if a user is logged in.
	useEffect(() => {
		if (user !== null && JwtService.getUserID() === null) {
			setUser(null)
		} else if (user === null && JwtService.getUserID() !== null) {
			setUser({})
			axiosInstance
				.get("/user/" + JwtService.getUserID())
				.then((response) => {
					if (response.data.userId) {
						setUser(response.data.profile)
						setLoading(false)
					}
				})
				.catch(() => {
					JwtService.rmUser()
				})
		}
	})

	return (
		<Navbar bg="dark" variant="dark" expand={false} sticky="top">
			<Container fluid>
				{/* Navbar toggle and brand title. */}
				<Nav className="flex-row me-auto">
					<Navbar.Toggle
						aria-controls="offcanvasNavbar"
						style={{ marginRight: "1rem" }}
					/>
					<HeadOffCanvas />
					<LinkContainer to="/">
						<Navbar.Brand>Project_Code</Navbar.Brand>
					</LinkContainer>
				</Nav>
				<LSearchBar />
				<XSSearchBar />

				{user !== null ? ( 
					!loading ? ( <LoggedInDropdown user={user} />) : ( <></>)) 
				: ( <LoggedOutForm />)}
			</Container>
		</Navbar>
	)
}

export default Navigation
