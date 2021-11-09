// import React from "react";
// import { Link, withRouter } from "react-router-dom";


// function navButton(props) {
// }

// function Navigation(props) {
// 	return (
// 		<div className="navigation">
// 			<nav class="navbar navbar-expand navbar-dark bg-dark">
// 				<div class="container">
// 					<Link class="navbar-brand" to="/">
// 						React Multi-Page Website
// 					</Link>
// 					<div>
// 						<ul class="navbar-nav ml-auto">
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/">
// 									Home
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/about">
// 									About
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/code">
// 									Code
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/contact">
// 									Contact	
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/login">
// 									Login
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 							<li
// 								class="nav-item"
// 							>
// 								<Link class="nav-link" to="/register">
// 									Register
// 									<span class="sr-only">(current)</span>
// 								</Link>
// 							</li>
// 						</ul>
// 					</div>
// 				</div>
// 			</nav>
// 		</div>
// 	)
// }

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