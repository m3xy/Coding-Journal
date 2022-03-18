import React from "react"
import { useDropzone } from "react-dropzone"
import styles from "./DragAndDrop.module.css"

const DragAndDrop = ({ handleDrop, children }) => {
	let { getRootProps, isDragActive } = useDropzone({
		handleDrop
	})

	return (
		<div className={styles.dragRoot} {...getRootProps()}>
			<input {...getRootProps()} />
			{isDragActive ? (
				<div className={styles.dragDiv0}>
					<div className={styles.dragDiv1}>
						<p>drop here</p>
					</div>
				</div>
			) : (
				<div></div>
			)}
			{children}
		</div>
	)
}

export default DragAndDrop
