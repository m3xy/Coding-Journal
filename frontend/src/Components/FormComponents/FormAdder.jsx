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
	setForm,
	validate
}) => {
	const val = useRef()
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")

	const valid = (required, val) => {
		return (!required && val.length === 0) || validate(elemName, val)
	}


	// Map elements of the array each to a button representing it.
	let cards = array.map((elem, i) => {
		return (
			<Button
				key={i}
				variant="outline-secondary"
				size="sm"
				className={styles.isolatedArrayButton}
				onClick={() => {
					setArray((array) => {
						return array.filter((value) => value !== elem)
					})
					setForm((form) => {
						return { ...form, [arrName]: array }
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
		setForm((form) => {
			return { ...form, [arrName]: array }
		})
		val.current.value = ""
		setInput("")
	}

	return (
		<Row>
			<Form.Label>{display}</Form.Label>
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
