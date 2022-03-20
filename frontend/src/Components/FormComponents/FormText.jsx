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
	as,
	onChange,
	isInvalid,
	placeholder,
	feedback,
	required,
	rows
}) => {
	return (
		<Form.Group className="mb-3" controlId={name}>
			<Form.Label>
				{" "}
				{display + (required ? "" : " (optional)")}{" "}
			</Form.Label>
			<Form.Control
				type={type ? type : "text"}
				as={as ? as : "input"}
				name={name}
				placeholder={
					placeholder ? placeholder : "Insert " + display + " here..."
				}
				isInvalid={isInvalid}
				rows={rows ? rows : 1}
				onChange={onChange}
			/>
			<Form.Control.Feedback type="invalid">
				{feedback ? feedback : ""}
			</Form.Control.Feedback>
		</Form.Group>
	)
}

export default FormText
