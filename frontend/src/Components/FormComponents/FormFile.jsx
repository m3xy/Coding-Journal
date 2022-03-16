import React, { useEffect, useState } from "react"
import Dropzone from "react-dropzone"
import { Form, ListGroup, Ratio, Card, CloseButton } from "react-bootstrap"
import styles from "./FormComponents.module.css"

const FileLabel = ({ children }) => {
	return <label style={{ margin: "5px" }}>{children}</label>
}
const FormFile = ({
	display,
	label,
	accept,
    placeholder,
	elemName,
	name,
	fileLimit,
	validate,
	setForm
}) => {
	const [files, setFiles] = useState([])

	useEffect(() => {
		setForm((form) => {
			return { ...form, [name]: files }
		})
	}, [files])

	// Handle dropping new files to the drag n' drop.
	const onDrop = (acceptedFiles) => {
		if (fileLimit ? acceptedFiles.length > fileLimit : false) return
		let success = acceptedFiles.map((file) => {
			let state = validate(elemName, file.hasOwnProperty('path'))
		})
		if (success.includes(false)) {
			return
		} else {
			setFiles(acceptedFiles)
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
		<Dropzone onDrop={onDrop}>
			{({ getRootProps, getInputProps }) => {
				return (
                <section className={styles.DropContainer} {...getRootProps()}>
                    <input {...getInputProps()} />
                    <p>
                        {placeholder
                            ? placeholder
                            : "Drop " + display + " here..."}
                    </p>
				</section>)
			}}
		</Dropzone>
	)

	return (
		<Form.Group controlId={label} className="mb-3">
			<Form.Label>{display}</Form.Label>
			<Form.Control
				type="file"
				accept={accept}
				onChange={(e) => {
                    console.log(e);
					onDrop([...e.target.files])
				}}
			/>
			<Card.Body>{files.length !== 0 ? cards : dropArea}</Card.Body>
		</Form.Group>
	)
}

export default FormFile
