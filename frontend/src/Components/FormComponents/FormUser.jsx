/*
 * FormUser.jsx
 * Reusable component for user query and selection.
 * Author: 190014935
 *
 * @param display The field's display name
 * @param name The name on the validation function
 * @param immutables users given from outside the form
 * which cannot be modified
 * @param initUsers Initial users edited from outside the
 * form which count in the array
 * @param query Base query made to get user
 * @param onChange function called on array change
 * CAUTION: array received as {target: {name: name, value: array}}
 */

import React, { useState, useEffect, useRef } from "react"
import { Form, Table, Button, Dropdown, InputGroup } from "react-bootstrap"
import axiosInstance from "../../Web/axiosInstance"

export default ({ display, name, immutables, initUsers, query, onChange }) => {
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")
	const [results, setResults] = useState([])
	const searchBar = useRef()

	// Send array features into the array.
	useEffect(() => {
		onChange({ target: { name: name, value: array } })
	}, [array])

	// Make a query based on the input to get user results.
	useEffect(() => {
		setResults([])
		if (input.length > 1)
			axiosInstance
				.get("/users/query", { params: { ...query, name: input } })
				.then((response) => {
					setResults(
						response.data.users !== undefined
							? response.data.users
							: []
					)
				})
				.catch(() => {
					return
				})
	}, [input])

	// Add initial users (added outside of this form control)
	// into the array.
	useEffect(() => {
		if (initUsers?.length > 0)
			setArray((array) => {
				return [
					...array.filter((elem) => {
						return !initUsers
							.map((initUser) => initUser.userId)
							.includes(elem.userId)
					}),
					...initUsers
				]
			})
	}, [initUsers])

	const handleChange = (e) => {
		setInput(e.target.value)
	}

	// Handle a user selection.
	const handleSelect = (user) => {
		setArray((array) => {
			return [...array, user]
		})
		setInput("")
		searchBar.current.value = ""
	}

	// Get a row entry for an user
	const getUserRow = (user, key, options) => {
		return (
			<tr key={key}>
				<td>{key}</td>
				<td>
					{user.firstName} {user.lastName}
				</td>
				<td>{user.profile?.email}</td>
				<td>
					{options ? (
						options
					) : (
						<div className="text-muted">Cannot modify...</div>
					)}
				</td>
			</tr>
		)
	}

	return (
		<div>
			<Form.Label>{display}</Form.Label>
			<InputGroup>
				<Form.Control
					placeholder={display}
					aria-label={display}
					onChange={handleChange}
					ref={searchBar}
				/>
			</InputGroup>
			<Dropdown
				show={
					input.length > 2 &&
					document.activeElement === searchBar.current
				}>
				<Dropdown.Menu>
					{results !== null ? (
						results?.map((resUser, i) => {
							return (
								<Dropdown.Item
									key={i + 1}
									onClick={() => {
										handleSelect(resUser)
									}}>
									{resUser.firstName} {resUser.lastName} {"("}
									{resUser.profile?.email}
									{")"}
								</Dropdown.Item>
							)
						})
					) : (
						<Dropdown.Item>
							<div className="text-muted">
								No users fit search...
							</div>
						</Dropdown.Item>
					)}
				</Dropdown.Menu>
			</Dropdown>
			<Table striped bordered hover style={{ marginTop: "15px" }}>
				<thead>
					<tr>
						<th>#</th>
						<th>Name</th>
						<th>Email</th>
						<th>Options</th>
					</tr>
				</thead>
				<tbody>
					{immutables?.map((user, i) => {
						return getUserRow(user, i)
					})}
					{array.map((user, i) => {
						return getUserRow(
							user,
							i + immutables?.length,
							<Button
								variant="danger"
								onClick={() => {
									setArray((array) => {
										return array.filter(
											(check) =>
												check.userId !== user.userId
										)
									})
								}}>
								Delete
							</Button>
						)
					})}
				</tbody>
			</Table>
		</div>
	)
}
