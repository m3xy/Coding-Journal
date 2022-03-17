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
	feedback,
	onChange,
	validate
}) => {
	const val = useRef()
	const [array, setArray] = useState([])
	const [input, setInput] = useState("")
	const [modded, setModded] = useState(false)

	useEffect(() => {
		if (modded) onChange({ target: { name: arrName, value: array } })
	}, [array, modded])

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
						setModded(true)
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
			<Form.Control.Feedback>
				{feedback ? feedback : ""}
			</Form.Control.Feedback>
		</Row>
	)
}

export default FormAdder
