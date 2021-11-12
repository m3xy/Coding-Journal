import React from "react";
import { Form, Button, FloatingLabel } from "react-bootstrap";

class Login extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            email: "",
            password: "",
            journal: ""
        }

        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleChange = this.handleChange.bind(this);
    }

    handleSubmit(e) {
        e.preventDefault();
        this.props.login(this.state.email, this.state.password, this.state.journal);

    }

    handleChange(e) {
        const { name, value } = e.target;
        this.setState({
            [name]: value
        });
    }

    render() {

        return(
            <div className="col-md-3 offset-4">
                <br/>
                <h2>Login</h2>
                <Form onSubmit={this.handleSubmit}>
                    <Form.Group className="mb-3" controlId="email">
                        <Form.Label>Email address</Form.Label>
                        <Form.Control type="email" placeholder="Enter email" name="email" onChange={this.handleChange} required/>
                        <Form.Text className="text-muted">
                        We'll never share your email with anyone else.
                        </Form.Text>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="password">
                        <Form.Label>Password</Form.Label>
                        <Form.Control type="password" placeholder="Password" name="password" onChange={this.handleChange} required/>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="journal">
                        <Form.Label>Journal</Form.Label>
                        <Form.Select name="journal" onChange={this.handleChange} required>
                            <option value="">Select journal</option>
                            <option value="2">Journal 2</option>
                            <option value="5">Journal 5</option>
                            <option value="8">Journal 8</option>
                            <option value="11">Journal 11</option>
                            <option value="13">Journal 13</option>
                            <option value="17">Journal 17</option>
                            <option value="20">Journal 20</option>
                            <option value="23">Journal 23</option>
                            <option value="26">Journal 26</option>
                        </Form.Select>
                    </Form.Group> 
                    <br />
                    <Button variant="primary" type="submit">
                        Login
                    </Button>
                </Form>

            </div>
        );
    };
}

export default Login;