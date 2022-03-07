/**
 * HeadOffCanvas.jsx
 * Author: 190014935
 *
 * Off canvas body for the header/navbar.
 */
import React from "react"
import {
	Navbar,
	Offcanvas,
	Nav
} from "react-bootstrap"
import { LinkContainer } from "react-router-bootstrap"

const HeadOffCanvas = () => {
	return (
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
	)
}
export default HeadOffCanvas
