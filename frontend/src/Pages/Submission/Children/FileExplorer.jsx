import React, { useEffect, useState } from "react"
import FileBrowser, { Icons } from "react-keyed-file-browser"
import { Card } from "react-bootstrap"
import "font-awesome/css/font-awesome.min.css"

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
				if (file.path.slice(-1) !== "/")
					// Check if given file is not a directory.
					struct = [
						...struct,
						{
							key: file.path,
							fileId: file.ID
						}
					]
			})
			setFiles(struct)
		}
	}

	return (
		<Card body>
			<h4>File Explorer</h4>
			<FileBrowser
				files={fileArray}
				icons={Icons.FontAwesome(4)}
				onSelect={(file) => {
					if (file !== undefined) onClick(file.fileId)
				}}
				canFilter={false}
				detailRenderer={() => null}
			/>
		</Card>
	)
}
