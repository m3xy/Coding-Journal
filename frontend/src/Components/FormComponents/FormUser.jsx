import React, { useState, useEffect, useRef, useCallback } from "react"
import { Form, Table, Button, Dropdown, InputGroup } from "react-bootstrap"
import axiosInstance from "../../Web/axiosInstance"

export default ({ display, name, immutables, initUsers, query, onChange }) => {
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")
	const [results, setResults] = useState([])
	const searchBar = useRef()

	useEffect(() => {
		onChange({ target: { name: name, value: array } })
	}, [array])

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
					{immutables.map((user, i) => {
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
