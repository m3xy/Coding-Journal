/*
 * Abstract.jsx
 * Card rendering the markdown for the abstract.
 * Author: 190014935
 */
import React from "react"
import ReactMarkdown from "react-markdown"
import { Card, Collapse, Button } from "react-bootstrap"

export default ({ markdown, show, setShow, inversed }) => {
	let showAbstract = inversed ? !show : show
	return (
		<Card style={{ marginBottom: "15px" }} body>
			<Collapse in={showAbstract}>
				<div id="collapse-abstract">
					<h4>Abstract</h4>
					{markdown !== "" ? (
						<ReactMarkdown children={markdown} />
					) : (
						<p className="text-muted">No abstract...</p>
					)}
				</div>
			</Collapse>
			<div style={{ display: "flex" }}>
				<Button
					style={{ flex: 6 }}
					variant="outline-secondary"
					onClick={() => setShow((show) => !show)}>
					{!showAbstract ? "Show abstract" : "Show files"}
				</Button>
			</div>
		</Card>
	)
}
