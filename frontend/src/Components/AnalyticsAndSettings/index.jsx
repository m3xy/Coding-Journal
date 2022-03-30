import React, { useState, useEffect } from "react"
import { Card, Button, Badge, Container, Alert, Tab, Tabs, Form, Row, Col} from "react-bootstrap"
import { useNavigate } from "react-router-dom"
import axios from "axios"
import { CSSTransition } from 'react-transition-group';
import axiosInstance from "../../Web/axiosInstance"
import DragAndDrop from "../DragAndDrop/index"
import { createMemoryHistory } from "history"

export default(id) => {
    const [form, setForm] = useState({
        firstName: null,
        lastName: null,
        phoneNumber: null,
        organization: null,
        password: null,
        repeatPassword: null

    })
    const [errors, setErrors] = useState({})
    const [show, setShow] = useState(false)
	const [alertMsg, setAlertMsg] = useState("")
    const [error, setError] = useState(null)
    // useEffect(() => {
            

            

        // }, []) 

    const fetchAnalytics = (id) => {

    }

    const deleteProfile = (id) => {
        axiosInstance
			.post("/user/" + id + "/delete")
			.then((response) => {
				console.log(response)
			})
			.catch((error) => {
				console.log(error)
			})
    }

    const updateProfile = (firstName, lastName, phoneNumber, organization, password) => {
        let data = {
            firstName: firstName,
            lastName: lastName,
            phoneNumber: phoneNumber,
            organization: organization,
            password: password
        }
        axiosInstance
        .post("/user/" + id + "/edit", data)
        .then((response) => {
            console.log(response)
        })
        .catch((error) => {
            console.log(error)
        })
    }

    const validate = (target, value) => {
		switch (target) {
			case "firstName":
                return value.length >= 0
			case "lastName":
				return value.length >= 0
			case "email":
				if (value.length == 0) {
					return true
				} else {
					return String(value)
						.toLowerCase()
						.match(
							/^(([^<>()\\[\]\\.,;:\s@"]+(\.[^<>()\\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
						) // From https://emailregex.com/
				}
			case "password":
				return (
					value.length > 0 &&
					String(value).match(
						/^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*?[#?!@$ %^&*-]).{8,}$/
					) // From https://ihateregex.io/expr/password
				)
			case "repeatPassword":
				return value === form.password
		}
	}
    
    const createAnalyticsAndSettings = (id) => {
        const [showButton, setShowButton] = useState(true);
        const [showMessage, setShowMessage] = useState(false);
        const [key, setKey] = useState('analytics');
            return (
                <Container style={{ paddingTop: '2rem' }}>
                <Card style={{ width: '40rem' }}>
            <Card.Body>
            <Card.Title>John Smith</Card.Title>
            <Card.Subtitle>jsmith@gmail.com</Card.Subtitle>
            <Card.Body>
                <Tabs
                id="controlled-tab-example"
                activeKey={key}
                onSelect={(k) => setKey(k)}
                className="mb-3"
                >
                <Tab eventKey="analytics" title="Analytics">
                    <Card
                    border="primary"
                    style={{ width: '18rem' }}
                    >
                    <Card.Header>
                        Total Submissions
                    </Card.Header>
                    <Card.Body>
                        <Card.Title>10</Card.Title>
                    </Card.Body>
                    </Card>
                    <br />
                    <Card
                    border="success"
                    style={{ width: '18rem' }}
                    >
                    <Card.Header>
                        Accepted Submissions
                    </Card.Header>
                    <Card.Body>
                        <Card.Title>5</Card.Title>
                    </Card.Body>
                    </Card>
                    <br />

                    <Card
                    border="warning"
                    style={{ width: '18rem' }}
                    >
                    <Card.Header>
                        Submissions Under-Review
                    </Card.Header>
                    <Card.Body>
                        <Card.Title>3</Card.Title>
                    </Card.Body>
                    </Card>
                    <br />
                    <Card
                    border="danger"
                    style={{ width: '18rem' }}
                    >
                    <Card.Header>
                        Rejected Submissions
                    </Card.Header>
                    <Card.Body>
                        <Card.Title>2</Card.Title>
                    </Card.Body>
                    </Card>
                    <br />
                </Tab>
                <Tab eventKey="edit" title="Edit Profile">
                    <Card.Body>
                    <Form>
                        <Row>
                        <Form.Group
                            as={Col}
                            controlId="formGridEmail"
                        >
                            <Form.Label>First Name</Form.Label>
                            <Form.Control
                            type="email"
                            placeholder="Change First Name"
                            />
                        </Form.Group>

                        <Form.Group
                            as={Col}
                            controlId="formGridPassword"
                        >
                            <Form.Label>Last Name</Form.Label>
                            <Form.Control
                            type="text"
                            placeholder="Change Last Name"
                            />
                        </Form.Group>
                        </Row>
                        <Row>
                        <Form.Group
                            as={Col}
                            controlId="formGridEmail"
                        >
                            <Form.Label>
                            Organization
                            </Form.Label>
                            <Form.Control
                            type="text"
                            placeholder="Change Organization"
                            />
                        </Form.Group>

                        <Form.Group
                            as={Col}
                            controlId="formGridPassword"
                        >
                            <Form.Label>
                            Phone Number
                            </Form.Label>
                            <Form.Control
                            type="text"
                            placeholder="Change Phone Number"
                            />
                        </Form.Group>
                        </Row>
                        <Row>
                        <Col>
                            <Button variant="danger">
                            Delete Account
                            </Button>
                        </Col>{' '}
                        <Col>
                            <Button variant="primary">
                            Save Changes
                            </Button>{' '}
                        </Col>{' '}
                        </Row>
                    </Form>
                    </Card.Body>
                </Tab>
                </Tabs>
            </Card.Body>
            </Card.Body>
        </Card>

        <CSSTransition
            in={showMessage}
            timeout={300}
            classNames="alert"
            unmountOnExit
            onEnter={() => setShowButton(false)}
            onExited={() => setShowButton(true)}
        >
            <Alert
            variant="primary"
            dismissible
            onClose={() => setShowMessage(false)}
            >
            <Alert.Heading>
                Animated alert message
            </Alert.Heading>
            <p>
                This alert message is being transitioned in and
                out of the DOM.
            </p>
            <Button onClick={() => setShowMessage(false)}>
                Close
                </Button>
                </Alert>
            </CSSTransition>
            </Container>
        )
    }

    return (
         <div>
             <div>   
                {error 
                ? "Something went wrong, please try again later..."
                :createAnalyticsAndSettings(0)
                }
            </div>
        </div>
    )
}