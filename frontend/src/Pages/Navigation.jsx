
// export default Navigation;
import React from 'react'
import { Link, withRouter } from "react-router-dom";
import { Navbar,Nav, NavLink , Container} from 'react-bootstrap';
import {Helmet} from "react-helmet";




class Navigation extends React.Component{

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
}

export default Navigation;
