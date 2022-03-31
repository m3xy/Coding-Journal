/*
 * FormAdder.jsx
 * Form component to add simple strings into an array.
 * Author: 190014935
 *
 * @param display The field's display name
 * @param elemName Single element name for validation function
 * @param arrName Array name for validation function
 * @param placeholder Placeholder text on form control
 * @param required Whether this field is required or not
 * @param onChange function to call on array change, for use on the form
 * CAUTION: array returned as {target: {name: arrName, value: array}}
 * @param validate The validation function
 */

import React, { useState, useRef } from "react"
import { useEffect } from "react"
import { Form, Button, InputGroup, Row, Col } from "react-bootstrap"
import styles from "./FormComponents.module.css"

// Array adder for the form.
const FormAdder = ({
	display,
	elemName,
	arrName,
	label,
	placeholder,
	required,
	onChange,
	validate
}) => {
	const val = useRef()
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")

	useEffect(() => {
		onChange({ target: { name: arrName, value: array } })
	}, [array])

	const valid = (required, val) => {
		return (
			(!required && val.length === 0) ||
			(validate(elemName, val) && !array.includes(input))
		)
	}

	// Map elements of the array each to a button representing it.
	let cards = array.map((elem, i) => {
		return (
			<Button
				key={i}
				variant="outline-danger"
				size="sm"
				className={styles.isolatedArrayButton}
				onClick={() => {
					setArray((array) => {
						return array.filter((value) => value !== elem)
					})
				}}>
				{elem}
			</Button>
		)
	})

	const handleChange = (e) => {
		setInput(e.target.value)
	}

	// Handle click on adding an element to the array
	const handleClick = () => {
		if (
			input.length <= 0 ||
			!validate(elemName, input) ||
			array.includes(input)
		) {
			return
		}
		setArray((array) => {
			return [...array, input]
		})
		val.current.value = ""
		setInput("")
	}

	return (
		<Row>
			<Form.Label>
				{" "}
				{display + (required ? "" : " (optional)")}{" "}
			</Form.Label>
			<InputGroup>
				<Form.Control
					placeholder={
						placeholder
							? placeholder
							: "Add " + display + " here..."
					}
					aria-label={display}
					aria-describedby={label}
					ref={val}
					onChange={handleChange}
					isInvalid={!valid(false, input)}
				/>
				<Button
					variant="outline-secondary"
					id={label}
					onClick={handleClick}
					disabled={!valid(true, input)}>
					Add
				</Button>
			</InputGroup>
			<Col>{cards}</Col>
		</Row>
	)
}

export default FormAdder
