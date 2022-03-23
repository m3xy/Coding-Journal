import React from "react"
import { Collapse, Card } from "react-bootstrap"

export default ({ id, show }) => {
	return (
		<Card>
			<Card.Body>
				<h4>File Viewer</h4>
				<Collapse in={show}>
					<p>Hey there {id}!</p>
				</Collapse>
			</Card.Body>
		</Card>
	)
}
