import React from "react"
import { Collapse, Card } from "react-bootstrap"
import Code from "./Code"

export default ({ id, show }) => {
	return (
		<Card>
			<Card.Body>
				<h4>File Viewer</h4>
				<Code id={id} show={show} />
			</Card.Body>
		</Card>
	)
}
