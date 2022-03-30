/*
 * XsSearchBar.jsx
 * Author: 190014935
 *
 * Navbar search bar modal for small screens.
 */

import React, { useState } from "react"
import {
	Nav,
	Modal,
	Form,
	FormControl,
	Button
} from "react-bootstrap"
import { BsSearch } from "react-icons/bs"
import { createSearchParams, useNavigate } from "react-router-dom"

const searchURL = "/search"

// Modal used for small screen searches
const SearchForm = () => {

	const [term, setTerm] = useState("")
	const navigate = useNavigate()

	return(
		<Form className="d-flex" onSubmit={(e) => {
			e.preventDefault()

			navigate({
				pathname: searchURL,
				search: createSearchParams({
					name: term
				}).toString()
			})
		}}>
			<FormControl
				type="search"
				placeholder="Submissions, authors, tags..."
				className="me-2 "
				aria-label="Search"
				onChange={(e) => {
					setTerm(e.target.value)
				}}
				value={term}
			/>
			<Button className="" variant="outline-secondary" type="submit">
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
