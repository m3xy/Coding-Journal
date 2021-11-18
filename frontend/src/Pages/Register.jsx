import React from 'react';
import { Form, Button } from 'react-bootstrap';
import axiosInstance from "../Web/axiosInstance";

const registerEndpoint = '/register'

class Register extends React.Component {

    constructor(props){
        super(props);

        this.state = {

            firstName: "",
            lastName: "",
            email: "",
            password: ""
        }

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);

    }

    /**
     * Sends a POST request to the go server to register a new user
     *
     * @param userFirstName the first name of the user being registered
     * @param userLastName the last name of the user being registered
     * @param userEmail the user's email (only one account per email)
     * @param userPassword the user's password
     */
    registerUser(userFirstName, userLastName, userEmail, userPassword) {
        // constructs JSON data to send to the backend
        let data = {
            firstname: userFirstName,
            lastname: userLastName,
            email: userEmail,
            password: userPassword
        };

        // Send register request to backend.
        axiosInstance.post(registerEndpoint, data)
                     .then(() => {
                         history.push('/')
                     })
                     .catch((error) => {
                         console.log(error)
                     })
    }

    handleSubmit(e) {
        e.preventDefault();
        this.registerUser(this.state.firstName, this.state.lastName, this.state.email, this.state.password);
    }

    handleChange(e) {
        const {name, value} = e.target;
        this.setState({
            [name]: value
        });
    }
    
    render() {
        return(
            <div className="col-md-3 offset-4">
                <br />
                <h2>Register</h2>
                <Form onSubmit={this.handleSubmit}>
                    <Form.Group className="mb-3" controlId="firstName">
                        <Form.Label>First Name</Form.Label>
                        <Form.Control type="text" name="firstName" placeholder="Enter first name" onChange={this.handleChange} required/>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="lastName">
                        <Form.Label>Last Name</Form.Label>
                        <Form.Control type="text" name="lastName" placeholder="Enter last name" onChange={this.handleChange} required/>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="email">
                        <Form.Label>Email address</Form.Label>
                        <Form.Control type="email" name="email" placeholder="Enter email" onChange={this.handleChange} required/>
                        <Form.Text className="text-muted">
                        We'll never share your email with anyone else.
                        </Form.Text>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="password">
                        <Form.Label>Password</Form.Label>
                        <Form.Control type="password" name="password" placeholder="Password" onChange={this.handleChange} required/>
                    </Form.Group>

                    <Button variant="primary" type="submit">
                    Register
                    </Button>
                </Form>

            </div>
        )

    }
}

export default Register;
