import React, { useState, useEffect, useRef } from "react"
import { Form, Table, Button, Dropdown, InputGroup } from "react-bootstrap"
import axiosInstance from "../../Web/axiosInstance"

export default ({ display, query, onChange }) => {
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
				.catch((err) => {
					return
				})
	}, [input])

	const handleChange = (e) => {
		setInput(e.target.value)
	}

	const handleSelect = (user) => {
		setArray((array) => {
			return [...array, user]
		})
		setInput("")
		val.current.value = ""
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
		</Row>
	)
}
