import React from 'react';
import { Form, Button } from 'react-bootstrap';

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

    handleSubmit(e) {
        e.preventDefault();
        this.props.register(this.state.firstName, this.state.lastName, this.state.email, this.state.password);
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