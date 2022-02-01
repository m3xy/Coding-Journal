
/**
 * Navigation.jsx
 * Author: 190010714
 * 
 * Navigation Bar Component Class
 */
import React from 'react'
import { Navbar,Nav, /*NavLink, */  Container} from 'react-bootstrap'
import { LinkContainer } from 'react-router-bootstrap'
// import { useNavigate } from 'react-router-dom'
// import {Helmet} from "react-helmet";




/*class Navigation extends React.Component{

    render(){
        return(
			
			<Navbar bg="dark" variant="dark" expand="lg">
				<Container>
				<Navbar.Brand href="/">Journal</Navbar.Brand>
				<Nav className="me-auto">
				<Nav.Link href="/">Home</Nav.Link>
				<Nav.Link href="/register">Register</Nav.Link>
				<Nav.Link href="/code">Code</Nav.Link>
				<Nav.Link href="/about">About</Nav.Link>
				<Nav.Link href="/contact">Contact</Nav.Link>
				<Nav.Link href="/commentModal">Comment</Nav.Link>
				<Nav.Link href="/upload">Upload</Nav.Link>
				<Nav.Link href="/profile">Profile</Nav.Link>

				</Nav>
				</Container>
			</Navbar>
			
        )  
    }
}*/

function Navigation() {
	return(

		<Navbar bg="dark" variant="dark" expand="lg">
			<Container>
				<LinkContainer to="/">
					<Navbar.Brand >Journal</Navbar.Brand>
				</ LinkContainer>
				<Nav className="me-auto">
					<LinkContainer to="/">
						<Nav.Link>Home</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/register">
						<Nav.Link >Register</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/code">
						<Nav.Link>Code</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/about">
						<Nav.Link>About</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/contact">
						<Nav.Link>Contact</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/commentModal">
						<Nav.Link>Comment</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/upload">
						<Nav.Link>Upload</Nav.Link>
					</ LinkContainer>
					<LinkContainer to="/profile">
						<Nav.Link>Profile</Nav.Link>
					</ LinkContainer>
				</Nav>
			</Container>
		</Navbar>

	)
}

export default Navigation;
