import React, { useState, useEffect, useRef } from "react"
import { Form, Table, Button, Dropdown, InputGroup } from "react-bootstrap"
import axiosInstance from "../../Web/axiosInstance"

export default ({ display, initUsers, query, onChange }) => {
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")
	const [results, setResults] = useState([])
	const searchBar = useRef()
	const dropdown = useRef()

	useEffect(() => {
		onChange(array)
	}, [array])

	useEffect(() => {
		setResults([])
		if (input.length > 3)
			axiosInstance
				.get("/users/query", { params: { ...query, name: input } })
				.then((response) => {
					setResults(
						response.users !== undefined ? response.users : []
					)
				})
				.catch(() => {
					return
				})
	}, [input])

	const handleChange = (e) => {
		setInput(e.target.value)
	}

	// Handle a user selection.
	const handleSelect = (user) => {
		setArray((array) => {
			return [...array, user]
		})
		setInput("")
		val.current.value = ""
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
		<Row>
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
					results.length > 0 &&
					input.length > 0 &&
					document.activeElement === searchBar.current
				}>
				<Dropdown.Menu>
					{results.map((resUser) => {
						return (
							<Dropdown.Item
								key={i}
								onClick={() => {
									handleSelect(resUser)
								}}>
								{resUser.userId}
							</Dropdown.Item>
						)
					})}
				</Dropdown.Menu>
			</Dropdown>
			<Table striped bordered hover>
				<thead>
					<tr>
						<th>#</th>
						<th>Name</th>
						<th>Email</th>
						<th>Options</th>
					</tr>
				</thead>
				<tbody>
					{initUsers.map((user, i) => {
						return getUserRow(user, i)
					})}
					{array.map((user, i) => {
						return getUserRow(
							user,
							i + initUsers?.length,
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
		</Row>
	)
}
