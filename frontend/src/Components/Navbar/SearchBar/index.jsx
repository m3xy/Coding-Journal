/*
 * XsSearchBar.jsx
 * Author: 190014935
 *
 * Navbar search bar modal for small screens.
 */

import React, {useState} from "react"
import {
	Nav,
	Modal,
	Form,
	FormControl,
	Button
} from "react-bootstrap"
import { BsSearch } from "react-icons/bs"

// Modal used for small screen searches
const SearchForm = () => {
	return(
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
}

// Search bar for smaller screens 
export const XSSearchBar = () => {
	const [showSearch, setShowSearch] = useState(false)

	return (
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
					<Modal.Body><SearchForm /></Modal.Body>
				</Modal>
			</Nav>
	)
}

// Search bar for bigger screens
export const LSearchBar = () => {
	return (
		<Nav className="me-auto d-none d-lg-block"><SearchForm /></Nav>
	)
}
