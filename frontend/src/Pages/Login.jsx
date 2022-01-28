import React, { useState } from "react";
import { Form, Button/*, FloatingLabel*/ } from "react-bootstrap";
import axiosInstance from "../Web/axiosInstance"
import { useNavigate } from "react-router-dom"

const loginEndpoint = '/glogin'

function Login() {
    const [email, setEmail] = useState()
    const [password, setPassword] = useState()
    const [journal, setJournal] = useState(11)
    const navigate = useNavigate();

    function loginUser(email, password, journal) {
        let data = {
            email: email,
            password: password,
            groupNumber: journal
        }
        axiosInstance.post(loginEndpoint, data)
                     .then((response) => {
                         console.log(response);
                         document.cookie = "userId=" + response.data["userId"] + "; SameSite=Lax";
                         navigate('/');
                     })
                     .catch((error) => {
                         console.log(error.config)
                         console.log(error)
                     });
    }

    async function handleSubmit(e) {
        e.preventDefault();
        loginUser(email, password, journal)
    }

    return(
        <div className="col-md-3 offset-4">
            <br/>
            <h2>Login</h2>
            <Form onSubmit={handleSubmit}>
                <Form.Group className="mb-3" controlId="email">
                    <Form.Label>Email address</Form.Label>
                  <Form.Control type="email" placeholder="Enter email" name="email" onChange={(e) => setEmail(e.target.value)} required/>
                    <Form.Text className="text-muted">
                    We'll never share your email with anyone else.
                    </Form.Text>
                </Form.Group>

                <Form.Group className="mb-3" controlId="password">
                    <Form.Label>Password</Form.Label>
                  <Form.Control type="password" placeholder="Password" name="password" onChange={(e) => setPassword(e.target.value)} required/>
                </Form.Group>

                <Form.Group className="mb-3" controlId="journal">
                    <Form.Label>Journal</Form.Label>
                  <Form.Select name="journal" onChange={(e) => setJournal(e.target.value)} default="11" required>
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
    )
}

export default Login;
