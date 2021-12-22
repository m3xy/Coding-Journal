import React from "react";
import { Form, Button/*, FloatingLabel*/ } from "react-bootstrap";
import axiosInstance from "../Web/axiosInstance"
// import cookiesInstance from '../Web/CookieHandler'
import { history } from '../_helpers';

const loginEndpoint = '/glogin'

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

    /**
     * Method to send login credentials to the backend to be verified
     *
     * @param userEmail the user's email. Each email can only have one account
     * @param userPassword the user's password as a string
     */
    loginUser(userEmail, userPassword, journal) {

        // constructs JSON data to send to the backend
        let data = {
            email: userEmail,
            password: userPassword,
            groupNumber:journal
        };

        // TODO - Remove comments once axios alternative fully works.
        // create a new XMLHttpRequest
        // var request = new XMLHttpRequest()

        // var response;

        // get a callback when the server responds
        // request.addEventListener('load', () => {
            // update the state of the component with the result here
            // console.log(request.responseText);
            // this.userId = request.responseText;

            // document.cookie = "userId=" + request.responseText + "; SameSite=Lax"; //Set a session cookie for logged in user
            // history.push('/');

            // TEMP: return response here, set the state of the login widget to be login approved
        // })
        // open the request and send request.
        // request.open('POST', BACKEND_ADDRESS + loginEndpoint)
        // this.sendSecureRequest(request, data)
        axiosInstance.post(loginEndpoint, data)
                     .then((response) => {
                         console.log(response);
                         this.userId = response.data["userId"];
                         // cookiesInstance.set('userId', response.data["userId"], {sameSite: 'lax', path: '/'});
                         document.cookie = "userId=" + response.data["userId"] + "; SameSite=Lax";
                         history.push('/');
                     })
                     .catch((error) => {
                         console.log(error.config)
                         console.log(error)
                     });

        // Send request with JSON data as body.
        // TEMP: return bool here to indicate success or failure#
    }


    handleSubmit(e) {
        e.preventDefault();
        this.loginUser(this.state.email, this.state.password, this.state.journal);
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
                        <Form.Select name="journal" onChange={this.handleChange} default="11" required>
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
