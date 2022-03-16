/*
 * FormText.jsx
 * Component used for form text inputs.
 * Author: 190014935
 */

import React from "react"
import { Form } from "react-bootstrap"

// Text input for the form.
const FormText = ({
	display,
	name,
	type,
	onChange,
	isInvalid,
	placeholder,
	rows
}) => {
	console.log(rows ? rows : 1)
	return (
		<Form.Group className="mb-3" controlId={name}>
			<Form.Label> {display} </Form.Label>
			<Form.Control
				type={type ? type : "text"}
				name={name}
				placeholder={
					placeholder ? placeholder : "Insert " + display + " here..."
				}
				isInvalid={isInvalid}
				rows={rows ? rows : 1}
				onChange={onChange}
			/>
		</Form.Group>
	)
}

export default FormText
