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
	FormControl,
	Button,
	Offcanvas,
	Modal,
	Dropdown
} from "react-bootstrap"
import { BsSearch } from "react-icons/bs"
import { MdArrowDropDownCircle } from "react-icons/md"
import { useNavigate } from "react-router-dom"
import { LinkContainer } from "react-router-bootstrap"
import JwtService from "../Web/jwt.service"
import axiosInstance from "../Web/axiosInstance"

function Navigation() {
	const [user, setUser] = useState(null)
	const [loading, setLoading] = useState(true)
	const [showSearch, setShowSearch] = useState(false)
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

	// Modal used for small screen searches
	let searchForm = (
		<Form className="d-flex">
			<FormControl
				type="search"
				placeholder="Submissions, authors, tags..."
				className="me-2 "
				aria-label="Search"
			/>
			<Button className="" variant="outline-secondary">
				<BsSearch />
			</Button>
		</Form>
	)

	// Toggle for custom user drop-down menu.
	const loginToggle = React.forwardRef(({ children, onClick }, ref) => (
		<a
			ref={ref}
			style={{ color: "#FFFFFF" }}
			onClick={(e) => {
				e.preventDefault()
				onClick(e)
			}}>
			{children} <MdArrowDropDownCircle />
		</a>
	))

	return (
		<Navbar bg="dark" variant="dark" expand={false} sticky="top">
			<Container fluid>
				{/* Navbar toggle and brand title. */}
				<Nav className="flex-row me-auto">
					<Navbar.Toggle
						aria-controls="offcanvasNavbar"
						style={{ marginRight: "1rem" }}
					/>
					<Navbar.Offcanvas>
						<Offcanvas.Header closeButton></Offcanvas.Header>
						<Offcanvas.Body>
							<Nav justify>
								<LinkContainer to="/">
									<Nav.Link>Home</Nav.Link>
								</LinkContainer>
								<LinkContainer to="/about">
									<Nav.Link>About</Nav.Link>
								</LinkContainer>
								<LinkContainer to="/contact">
									<Nav.Link>Contact</Nav.Link>
								</LinkContainer>
								<LinkContainer to="/upload">
									<Nav.Link>Publish</Nav.Link>
								</LinkContainer>
							</Nav>
						</Offcanvas.Body>
					</Navbar.Offcanvas>
					<LinkContainer to="/">
						<Navbar.Brand>Project_Code</Navbar.Brand>
					</LinkContainer>
				</Nav>

				{/* Search bar for bigger screens */}
				<Nav className="me-auto d-none d-lg-block">{searchForm}</Nav>

				{/* Search bar for smaller screens */}
				<Nav className="me-auto d-block d-lg-none">
					<Button
						variant="outline-secondary"
						onClick={() => {
							setShowSearch(true)
						}}>
						<BsSearch />
					</Button>
					<Modal
						show={showSearch}
						onHide={() => {
							setShowSearch(false)
						}}>
						<Modal.Header closeButton>
							<Modal.Title>
								Search submissions, authors, tags...
							</Modal.Title>
						</Modal.Header>
						<Modal.Body>{searchForm}</Modal.Body>
					</Modal>
				</Nav>

				{user !== null ? (
					/* Logged in component - Full Name with drop-down Menu */
					!loading ? (
						<Dropdown>
							<Dropdown.Toggle>
								{user.firstName + " " + user.lastName}
							</Dropdown.Toggle>
							<Dropdown.Menu variant="dark" align="end">
								<Dropdown.Item
									onClick={() => {
										navigate("/profile")
									}}>
									{" "}
									Profile{" "}
								</Dropdown.Item>
								<Dropdown.Item
									onClick={() => {
										navigate("/submissions")
									}}>
									{" "}
									Submissions{" "}
								</Dropdown.Item>
							</Dropdown.Menu>
						</Dropdown>
					) : (
						<></>
					)
				) : (
					/* Logged out component - Register and Log In buttons. */
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
				)}
			</Container>
		</Navbar>
	)
}

export default Navigation
