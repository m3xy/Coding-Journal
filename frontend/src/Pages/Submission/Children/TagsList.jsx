import React from "react"
import { Button, Card } from "react-bootstrap"

export default ({ tags }) => {
	return (
		<Card style={{ marginTop: "15px" }} body>
			<h4>Tags</h4>
			{tags !== undefined ? (
				tags.map((tag, i) => {
					return (
						<Button
							key={i}
							variant="outline-secondary"
							size="sm"
							disabled
							style={{ margin: "3px" }}>
							{tag.category}
						</Button>
					)
				})
			) : (
				<div className="text-muted">No tags...</div>
			)}
		</Card>
	)
}
