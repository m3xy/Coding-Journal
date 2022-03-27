/*
 * FileExplorer.jsx
 * Card showing the file explorer for the submission
 * Author: 190014935
 *
 * @param files The submission's file array.
 * @param onClick (file.fileId) Function handling a file ID given.
 */
import React, { useEffect, useState } from "react"
import FileBrowser, { Icons } from "react-keyed-file-browser"
import { Card } from "react-bootstrap"
import "font-awesome/css/font-awesome.min.css"
import styles from "./Submission.module.css"
import moment from "moment"

export default ({ files, onClick }) => {
	const [fileArray, setFiles] = useState([])

	useEffect(() => {
		fileArrayToKeyedStruct(files)
	}, [files])

	// Get a file array of correct format from the files given by the parent structure.
	const fileArrayToKeyedStruct = (array) => {
		if (array !== undefined) {
			let struct = []
			array.map((file) => {
				if (file.path.slice(-1) !== "/") {
					// Check if given file is not a directory.
					struct = [
						...struct,
						{
							key: file.path,
							fileId: file.ID,
							modified: +moment(file.CreatedAt)
						}
					]
				}
			})
			setFiles(struct)
		}
	}

	const onSelect = (file) => {
		if (file.key.slice(-1) !== "/") onClick(file.fileId)
	}

	return (
		<Card body>
			<h4>File Explorer</h4>
			<div className={styles.fileBrowser}>
				<FileBrowser
					files={fileArray}
					icons={Icons.FontAwesome(4)}
					onSelect={onSelect}
					canFilter={false}
					detailRenderer={() => null}
				/>
			</div>
		</Card>
	)
}
