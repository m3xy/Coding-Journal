import React, { useState, useRef } from "react"
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
	const val = useRef("")
	const [array, setArray] = useState([])

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

	// Handle click on adding an element to the array
	const handleClick = () => {
		let value = val.current.value
		if (
			value.length <= 0 ||
			!validate(elemName, value) ||
			array.includes(value)
		) {
			return
		}
		setArray((array) => {
			return [...array, value]
		})
		setForm((form) => {
			return { ...form, [arrName]: array }
		})
		val.current.value = ""
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
					isInvalid={array.includes(val.current)}
				/>
				<Button
					variant="outline-secondary"
					id={label}
					onClick={handleClick}
					disabled={
						!val.current.length === 0 || array.includes(val.current)
					}>
					Add
				</Button>
			</InputGroup>
			<Col>{cards}</Col>
		</Row>
	)
}

export default FormAdder
