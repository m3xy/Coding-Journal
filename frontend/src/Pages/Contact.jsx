import React from "react"
import { Form, Button, Row, Col, Container } from "react-bootstrap"

function Contact() {
	return (
		<Container>
			<Row>
				<Col></Col>
				<Col xs={4}>
					<h2 style={{marginTop:"20px", marginBottom:"20px"}}>Contact us!</h2>
					<Form>
						<Form.Group className="mb-3">
							<Form.Label>Name*</Form.Label>
							<Form.Control
								type="text"
								required
							/>
						</Form.Group>
						<Form.Group className="mb-3">
							<Form.Label>Email address*</Form.Label>
							<Form.Control
								type="email"
								required
							/>
						</Form.Group>
						<Form.Group className="mb-3">
							<Form.Label>Message*</Form.Label>
							<Form.Control as="textarea" rows={3} required />
						</Form.Group>
						<Button variant="outline-dark" type="submit">
							Submit
						</Button>
					</Form>
				</Col>

				<Col></Col>
			</Row>
		</Container>
	)
}

export default Contact
