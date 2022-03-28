/*
 * FormFile.jsx
 * Form component to select (or drag) files into the form.
 * Author: 190014935
 *
 * @param display The display name for the field.
 * @param label The DOM label of the field
 * @param accept List of file formats (as string separated with commas)
 * 		which the field accepts.
 * @param placeholder The message placeholder on the form control.
 * @param name name used on validation function for the whole array.
 * @param fileLimit Limit of files to add.
 * @param validate  The validation function ((name, value) => boolean)
 * @param feedback Default message to use on errors
 * @param onChange Function connecting form to this component.
 * 		CAUTION: added to onChange as {target: {name, value}} object.
 */

import React, { useEffect, useState } from "react"
import { useDropzone } from "react-dropzone"
import { Form, ListGroup, Card, CloseButton } from "react-bootstrap"
import styles from "./FormComponents.module.css"

const FileLabel = ({ children }) => {
	return <label style={{ margin: "5px" }}>{children}</label>
}
const FormFile = ({
	display,
	label,
	accept,
	placeholder,
	name,
	fileLimit,
	validate,
	feedback,
	onChange
}) => {
	const [files, setFiles] = useState([])
	const [modded, setModded] = useState(false)
	const { getRootProps, getInputProps } = useDropzone({
		onDrop: (files) => onDrop(files),
		accept: accept.split(",")
	})
	const [invalid, setInvalid] = useState(false)

	useEffect(() => {
		if (modded) onChange({ target: { name: name, value: files } })
	}, [files])

	// Handle dropping new files to the drag n' drop.
	const onDrop = (files) => {
		if (fileLimit ? files.length > fileLimit : false) return
		if (!validate(name, files)) {
			setInvalid(true)
			return
		} else {
			setInvalid(false)
			setModded(true)
			setFiles(files)
		}
	}

	let cards = files.map((file, i) => {
		return (
			<ListGroup.Item key={i}>
				<CloseButton
					onClick={() => {
						setFiles((files) => {
							return files.filter((rmFile) => {
								rmFile.name != file.name
							})
						})
					}}
				/>
				<FileLabel>File name: {file.name} </FileLabel>
				<FileLabel>File type: {file.type} </FileLabel>
				<FileLabel>File Size: {file.size} </FileLabel>
				<FileLabel>
					Last modified: {new Date(file.lastModified).toUTCString()}
				</FileLabel>
			</ListGroup.Item>
		)
	})

	let dropArea = (
		<section className={styles.DropContainer} {...getRootProps()}>
			<input {...getInputProps()} />
			<p>{placeholder ? placeholder : "Drop " + display + " here..."}</p>
		</section>
	)

	return (
		<Form.Group controlId={label} className="mb-3">
			<Form.Label>{display}</Form.Label>
			<Form.Control
				type="file"
				accept={accept}
				isInvalid={invalid}
				onChange={(e) => {
					console.log(e.target.files)
					onDrop([...e.target.files])
				}}
			/>
			<Form.Control.Feedback type="invalid">
				{feedback
					? feedback
					: "Please drop a file/files in the correct format."}
			</Form.Control.Feedback>
			<Card.Body>{files.length !== 0 ? cards : dropArea}</Card.Body>
		</Form.Group>
	)
}

export default FormFile
