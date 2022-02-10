/**
 * Register.jsx
 * Author: 190019931
 * 
 * This file stores the info for rendering the Register page of our Journal
 */

import React, { useState } from 'react';
import { Form, Button, Container, Row, Col } from 'react-bootstrap'
import axiosInstance from "../Web/axiosInstance"
import { useNavigate } from "react-router-dom"
import 'regenerator-runtime/runtime'

const registerEndpoint = '/register'

function Register() {
    const navigate = useNavigate()
    const [firstName, setFirstName] = useState()
    const [lastName, setLastName] = useState()
    const [email, setEmail] = useState()
    const [password, setPassword] = useState()

    function registerUser(firstName, lastName, email, password) {
        let data = {
            firstname: firstName,
            lastname: lastName,
            email: email,
            password: password
        }


        // Send register request to backend.
        axiosInstance.post(registerEndpoint, data)
                     .then(() => {
                         navigate('/');
                     })
                     .catch((error) => {
                         console.log(error);
                     })
    }

    async function handleSubmit(e) {
        e.preventDefault()
        registerUser(firstName, lastName, email, password)
    }

    return(
        <Container>
            <Row>
                <Col></Col>
                <Col xs={4}>
                    <br />
                    <h2>Register</h2>
                    <Form onSubmit={handleSubmit}>
                        <Form.Group className="mb-3" controlId="firstName">
                            <Form.Label>First Name</Form.Label>
                        <Form.Control type="text" name="firstName" placeholder="Enter first name" onChange={(e) => { setFirstName(e.target.value) }} required/>
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="lastName">
                            <Form.Label>Last Name</Form.Label>
                        <Form.Control type="text" name="lastName" placeholder="Enter last name" onChange={(e) => { setLastName(e.target.value) } } required/>
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="email">
                            <Form.Label>Email address</Form.Label>
                        <Form.Control type="email" name="email" placeholder="Enter email" onChange={(e) => { setEmail(e.target.value) }} required/>
                            <Form.Text className="text-muted">
                            We'll never share your email with anyone else.
                            </Form.Text>
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="password">
                            <Form.Label>Password</Form.Label>
                        <Form.Control type="password" name="password" placeholder="Password" onChange={(e) => {setPassword(e.target.value)} } required/>
                        </Form.Group>

                        <Button variant="primary" type="submit">
                        Register
                        </Button>
                    </Form>
                </Col>
                <Col></Col>
            </Row>
        </Container>
    )
}

export default Register;
