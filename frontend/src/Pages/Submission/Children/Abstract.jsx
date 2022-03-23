import React from "react"
import ReactMarkdown from "react-markdown"
import { Card, Collapse, Button } from "react-bootstrap"

export default ({ markdown, show, setShow, inversed }) => {
	let showAbstract = inversed ? !show : show
	return (
		<Card style={{ marginBottom: "15px" }}>
			<Collapse in={showAbstract}>
				<div id="collapse-abstract">
					<Card.Body>
						<h4>Abstract</h4>
						{markdown !== "" ? (
							<ReactMarkdown children={markdown} />
						) : (
							<p className="text-muted">No abstract...</p>
						)}
					</Card.Body>
				</div>
			</Collapse>
			<Card.Body>
				<div style={{ display: "flex" }}>
					<Button
						style={{ flex: 6 }}
						variant="outline-secondary"
						onClick={() => setShow((show) => !show)}>
						{!showAbstract ? "Show abstract" : "Show files"}
					</Button>
				</div>
			</Card.Body>
		</Card>
	)
}
