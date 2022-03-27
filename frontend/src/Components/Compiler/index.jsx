import React, { useState, useEffect } from "react"
import { Card, Button, Badge, Container, Alert, Tab, Tabs, Form} from "react-bootstrap"
import styles from "./Compiler.module.css"
import { useNavigate } from "react-router-dom"
import axios from "axios"
import { CSSTransition } from 'react-transition-group';
import axiosInstance from "../../Web/axiosInstance"
import DragAndDrop from "../DragAndDrop/index"

const baseURL = "https://judge0-ce.p.rapidapi.com/submissions" //use wait param to only use one tag!!!!

const params = {base64_encoded: 'true', wait: 'true', fields: '*'}
const headers = {
    'content-type': 'application/json',
    'Content-Type': 'application/json',
    'X-RapidAPI-Host': 'judge0-ce.p.rapidapi.com',
    'X-RapidAPI-Key': '488c3c10c9msh8121e54c23bb036p1afd03jsn6192663efef7'
}

//Single file data that accepts stdin
const singleFileData = {
    language_id: 52,
    source_code: 'I2luY2x1ZGUgPHN0ZGlvLmg+CgppbnQgbWFpbih2b2lkKSB7CiAgY2hhciBuYW1lWzEwXTsKICBzY2FuZigiJXMiLCBuYW1lKTsKICBwcmludGYoImhlbGxvLCAlc1xuIiwgbmFtZSk7CiAgcmV0dXJuIDA7Cn0=',
    stdin: 'SnVkZ2Uw'
}

const multiFileData ={
    language_id: 89,
    additional_files: "UEsDBBQAAAAAAAaMeVQAAAAAAAAAAAAAAAAEACAAc3JjL1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAUEsDBBQACAAIABqLeVQAAAAAAAAAAAQYAAANACAAc3JjLy5EU19TdG9yZVVUDQAH5fo9Yub6PWLW+z1idXgLAAEE9QEAAAQUAAAA7ZjNSsNAFIXPTYMMqDBLl7NxW+gbTEtduNS+gE3rpgQKarvOG/mIeidzqoE0C0Fo0fuF4QtMzswki/kJAJnt1hPAA3DIlnRzBMfSo6BHKdy2scESeyzv6+3qeFtnRxq7wxRrvcbd8Vf1tkL+MO9axh8tbeYKj3jGK3ao8aYeyMmFetLJXWOhmUozL/r8Sj3U44OW207SMAzDMH4PyXKXpx2GYRhnSJofAh3pJltYX9BlJ+PpQEe6yRY+V9Al7WhPBzrSTTYnLeHhQ9jz4fAing50/NErG8a/YZTl0/p/N3z+NwzjDyPlfDGf4etA0COttUHL0yEArubobwKK/LPwBt/1gY50k20bAcM4FZ9QSwcI2bWxrA0BAAAEGAAAUEsDBBQACAAIABqLeVQAAAAAAAAAAHgAAAAYACAAX19NQUNPU1gvc3JjLy5fLkRTX1N0b3JlVVQNAAfl+j1i5vo9Yun8PWJ1eAsAAQT1AQAABBQAAABjYBVjZ2BiYPBNTFbwD1aIUIACkBgDJxAbAbEbEIP4FUDMAFPhIMCAAziGhARBmRUwXegAAFBLBwgLiMA4NQAAAHgAAABQSwMEFAAIAAgAs4Z4VAAAAAAAAAAAoQAAAAwAIABzcmMvQWRkLmphdmFVVA0ABxOiPGLY+z1i1vs9YnV4CwABBPUBAAAEFAAAACsoTcrJTFZIzkksLlZwTElRqObiUgCCzLwShURrODMJwUyGMMFEWX5mikJiSooGSKJCQQesoFJBsxqiBEwmK9gCpbQVKiH6giuLS1Jz9fJLS/QKioDKc/I0lJyBSpSASpI1IWpqETbUAgBQSwcIAburoWoAAAChAAAAUEsDBBQACAAIAAaMeVQAAAAAAAAAAG4DAAANACAAc3JjL0FkZC5jbGFzc1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAjVNbbxJBFP4OFJbL2mJrRe0NlSqgFm9JH1CTStJIgpeEpj744sBO2i3LDNkdmvBP/BsaL0kf/AH+KOOZhaRVqnGTM+f2nW/Onpz58fPkO4BtPM7BwaUs0ljO4jKKixxdyWEVaw7WHWwQSLC0WLosPUL6ia9884yQrFT3CXNN7UnCQttX8tVo0JXhnugGHEkKzyOkKq2WheU6ehT25K5vU5kdz9s6EsfChYsLLvLIOii5uI4bhMJzrU1kQjF8Kc2h9qKFNG5mUHaxiVsObruooMr0O5a+YFnqgVAH9dfdI9kzv4U648jIAWP1iBPL7Tjj6/qb0FemY0IpBo0carjDfTZLT0tEKA5EXza16gnz1jeHbEVGKBNxeaVVbZ/hNkxy0CAsncNKcIbWC5Stm62q7ju4a3/oPqF8mvXVse7LKWjSxK7oGR2OHTzkybe17o+GBLellAybgYgiyZ19OHvDlGMyuxdCeYGMypPCxmwjfyvcGw/lOfB37T/n3ajOUjRFEHR8IxsOHhE2/6s3wvq/cbwcaV5VIMmSQRb2y011PtYJkN0nPufZW2NNrFO1b6CPbPCS8pmOg1YKuDiFbjOppS3WVlbff0HiE5K1z0icYA74itRp8XwMs08mz4/FZW8xvncJ9ThOuIcH2MJVXIm7JLYI134BUEsHCD+uFSf1AQAAbgMAAFBLAwQUAAgACAAGjHlUAAAAAAAAAACrAwAAEgAgAHNyYy9TdWJ0cmFjdC5jbGFzc1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAjVJdTxNBFD3Dtt12u7UIQkUQQQu0VanfmFR9sAmhSRWSEnzwxWk7gaXbmWZ3StJ/4t/QtJrw4A/wRxnvbGsggMZN7r0zc889e+7M/fnr9AeALbx0kMJ8CknkjLtp3EIKt7A4Q/k7DlawauOujXsMrEtWI5NkPYbEK096+g2DVSgeMMSqqi0YsnVPivf9blME+7zp00k67Dd1wFvaU1QZL9RqBu40VD9oiW3PQDKNCWTzmJ9wF9eQdeEg7cI1LoO0jbyLNawzTL9VSoeE7r0T+ki1w2wCG0kUXBRRsnHfxQM8ZEj+oaQKQ1r2uTws7zaPxYWjxiDUglqzVJ8Sc/Uo46nyXuBJ3dCB4N2Kg02UqcW9ldeMIdflHVFVssX1B08f0SrUXOqQqgu1Yv0ctSaOwwrD7BWkDHbP7Hxp6i5XFQ9sPDJNPWPIn2U9eaI6YgIai9imLlUwsPGCHqWuVKdPr+PWpBRB1edhKEjZ5/N/mHCM72+Hy7Yvwvy4sHJZyN8K9wc9cQX8Y/3idVeKlymq3PcbnhYVG1sMa/+ljWH53zis0vimaHAtMhofmM+dxEwUp8DMdJGfpt1tioxivPQd7AstGK6TT0SHxmYwO4HuEqmh3SgtjjBVWhrBKpWGtBrC+jRC7CuRDBE7pWJ8g31GNo84+ST5FGlykCMd66TKwo1IzxyeRmoZHuM5nmABSxGeReqWfwNQSwcIOHlncRUCAACrAwAAUEsDBBQACAAIABGGeFQAAAAAAAAAAL0AAAARACAAc3JjL1N1YnRyYWN0LmphdmFVVA0AB+OgPGLY+z1i1vs9YnV4CwABBPUBAAAEFAAAAEWMTQrCQAxG957io6sWpReQ3kHoCdrRxcD8hElGEPHujRNbs0heyJdHdQ3ewYWFGXNdpSxO8D5ByydBvB6Y/kiKbXlmfwf/3nxO/fdaLi3Eg2lai5hQTJAU2ZAUI867en6xPOKYq4xUVBFS390wdZqgwSIfU9rYAFBLBwj388OfeAAAAL0AAABQSwMEFAAIAAgABox5VAAAAAAAAAAAfQIAABEAIABzcmMvUmVzdWx0cy5jbGFzc1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAbVHbahsxFBz5JlvexJs292vT6zq3TZ7ykFIohcKCm0I2FELpg2wLo2DLYVdb2s9qH+ISQz4gHxVytHUJodkHjc7s0Wjm6Ob26hrAIQ4EprHAsVjDEpYFCljhWBUoO3JNgGPBkesCG3jm6E3HPed4wfGSofJWG23fMRSD5heG0odhVzE0Wtqo42zQVsmpbPeJKQ2kNgzzwdfWufwuw740vTC2iTa9I3dQxMMs6aiP2jV7JyrN+jbdc60eGvAZZvJjmdX9MO5IY1TC8crDa7xxDQEZeN/tMlTjrG0T2bEemtji2Pawg10Pewg97CNk4BNxBv/eyef2uerYB1T8M7VqwFBwvuf+utbDMDIXmSXjSg6OGBaDR3/kozDqBynOBs3/E/+Lk3ORsaqnEvJ+IZNUUUn3BY+MKaKQ0oUsB1Hkrqink7B6aLCJKXpJ9xXA3MxonaFqjZARlrdGYL9ow/CE1kpOFlHDU8xOWr9RXSI8GKNw9hvFEUqtMcpnI1Q+jcEJq8c7vti+RO0PxCXqu/66T+jdyy6TMFBFHYKE61ghVwH52s/9zOXu5u8AUEsHCIg24rKqAQAAfQIAAFBLAwQUAAgACAAjjHlUAAAAAAAAAAA3AQAAEAAgAHNyYy9SZXN1bHRzLmphdmFVVA0AB9L8PWLU/D1i0vw9YnV4CwABBPUBAAAEFAAAAE2OwWrDMAxA7/kKHZ1RTNodyw477tocRw+yI4JbRw6W0nWU/fu84LAaYVl6z0JNmOaUFS54Q7toiPbl2MyLi8GDjygCJ5IlqsCjaaBEORWLopZ0S2GACQObXnPg8fMMmEdpH6vbe2SmDFcHb8D0tTVM/y1Kkw3cHldxvd6HAZK7VLVUptJ+cZrR6x89bJNqb3MKsli+7LsdfLDSSNnOmIVKYa7OMt2L2/7bByt1REhsXrvdvquw2D4mIfO83M/T8xdQSwcI5x9rrMEAAAA3AQAAUEsDBBQACAAIAKKEeVQAAAAAAAAAAPQAAAAHACAAY29tcGlsZVVUDQAHsO89Ytj7PWLW+z1idXgLAAEE9QEAAAQUAAAAXY7BCoJAGITv/1P8mqdA7SxISEgeskS9Rei6rri1ueZaHcR3zzxlcxkGPmZmpdkFb+yCqBogDfxsd/CSJPLSwIUoPu1jL8yOXui7MVNP0SvrSl4EaImqo1DJDjnyBnOh0LJswQt7PRFdDjiplLMtW41hkUfHGPgIpWwYwLebokkFUaolfY265fzzOhq/vwB4hWc0tmiyB27wAn3NmnmX0VqiTuW95YLhW3Y3Vmo6VPwDUEsHCAQiqBOxAAAA9AAAAFBLAwQUAAgACACghXlUAAAAAAAAAAC4AAAAAwAgAHJ1blVUDQAHjPE9Ytj7PWLW+z1idXgLAAEE9QEAAAQUAAAAU1bUT8rM009KLM5QUOAK8XCNd/ZxDA4OcAzxsOUCM+P9HH1dbYNSi0tzSoq5klMUiouSudLyixQyFTLzFBJyihX09PRzMpP0tfSyEosSuBSAICUfTKEap1KNwq+1UqnOrOVKyc9L5eLKSixLVNBNzkksLi5ILMlQUNKzQleupKCCcA8XAFBLBwiNn3sIhQAAALgAAABQSwECFAMUAAAAAAAGjHlUAAAAAAAAAAAAAAAABAAgAAAAAAAAAAAA7UEAAAAAc3JjL1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAUEsBAhQDFAAIAAgAGot5VNm1sawNAQAABBgAAA0AIAAAAAAAAAAAAKSBQgAAAHNyYy8uRFNfU3RvcmVVVA0AB+X6PWLm+j1i1vs9YnV4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIABqLeVQLiMA4NQAAAHgAAAAYACAAAAAAAAAAAACkgaoBAABfX01BQ09TWC9zcmMvLl8uRFNfU3RvcmVVVA0AB+X6PWLm+j1i6fw9YnV4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIALOGeFQBu6uhagAAAKEAAAAMACAAAAAAAAAAAACkgUUCAABzcmMvQWRkLmphdmFVVA0ABxOiPGLY+z1i1vs9YnV4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIAAaMeVQ/rhUn9QEAAG4DAAANACAAAAAAAAAAAACkgQkDAABzcmMvQWRkLmNsYXNzVVQNAAec/D1infw9Ypz8PWJ1eAsAAQT1AQAABBQAAABQSwECFAMUAAgACAAGjHlUOHlncRUCAACrAwAAEgAgAAAAAAAAAAAApIFZBQAAc3JjL1N1YnRyYWN0LmNsYXNzVVQNAAec/D1infw9Ypz8PWJ1eAsAAQT1AQAABBQAAABQSwECFAMUAAgACAARhnhU9/PDn3gAAAC9AAAAEQAgAAAAAAAAAAAApIHOBwAAc3JjL1N1YnRyYWN0LmphdmFVVA0AB+OgPGLY+z1i1vs9YnV4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIAAaMeVSINuKyqgEAAH0CAAARACAAAAAAAAAAAACkgaUIAABzcmMvUmVzdWx0cy5jbGFzc1VUDQAHnPw9Yp38PWKc/D1idXgLAAEE9QEAAAQUAAAAUEsBAhQDFAAIAAgAI4x5VOcfa6zBAAAANwEAABAAIAAAAAAAAAAAAKSBrgoAAHNyYy9SZXN1bHRzLmphdmFVVA0AB9L8PWLU/D1i0vw9YnV4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIAKKEeVQEIqgTsQAAAPQAAAAHACAAAAAAAAAAAACkgc0LAABjb21waWxlVVQNAAew7z1i2Ps9Ytb7PWJ1eAsAAQT1AQAABBQAAABQSwECFAMUAAgACACghXlUjZ97CIUAAAC4AAAAAwAgAAAAAAAAAAAApIHTDAAAcnVuVVQNAAeM8T1i2Ps9Ytb7PWJ1eAsAAQT1AQAABBQAAABQSwUGAAAAAAsACwDqAwAAqQ0AAAAA",
    stdin: "MTA="
}

const config = {
    headers: headers,
    params: params
}


export default () => {
	const [submissionToken, setSubmissionToken] = useState(null)
    const [languageID, setLanguageID] = useState(null) //probably a let somewhere else
    const [multifileProgram, setMultiFileProgram] = useState(null) //if this is a mutli file program (boolean)
    const [submission, setSubmission] = useState(null) //source code if single file additional files if multi file
    const [stdInput, setStdInput] = useState(null) //if this program accepts stdin (boolean)
    const [commandLineArgs, setArgs] = useState(null) //if this program accepts command line arguments (boolean)
    const [inputFile, setInputFile] = useState(null)
    const [enableNetwork, setEnableNetwork] = useState(null) //if this program requires network access (boolean)
    const [output, setOutput] = useState(null) 
    const [runTime, setRunTime] = useState(null) 
    const [memoryUsage, setMemoryUsage] = useState(null) 
    const [isLoading, setLoading] = useState(true);
    const [error, setError] = useState(null)
	const navigate = useNavigate()

    // useEffect(() => {
		

        

    // }, []) 

    const runCode = (submission) =>{
        axios.post(baseURL, multiFileData, config)
           .then(function (response) {
             console.log(response);
             setSubmissionToken(response.data.token)
             setOutput(atob(response.data.stdout))
             setMemoryUsage(response.data.memory)
             setRunTime(response.data.time)
             console.log("Response:" + response)
             console.log("Output: " + output)
             setLoading(false)
           })
           .catch(function (error) {
             console.log(error);
             setLoading(false)
           });

    }
    
	const createSubmission = (submission) => {
        const [showButton, setShowButton] = useState(true);
        const [showMessage, setShowMessage] = useState(false);
        const [showMemory, setShowMemory] = useState(false);
        const [showTime, setShowTime] = useState(false);
        const [key, setKey] = useState('run');
                return (
                <Container>
                    <Card
                        style={{
                            minWidth: "35rem",
                            maxWidth: "25rem",
                            margin: "8px"
                        }}
                        className="shadow rounded">
                        <Tabs
                            id="controlled-tab-example"
                            activeKey={key}
                            onSelect={(k) => setKey(k)}
                            className="mb-3"
                            >
                            <Tab eventKey="run" title="Run">
                                
                                <Card.Body>
                                    <Card.Title>{"This Code Can Be Run"}</Card.Title>
                                    <Card.Subtitle className="mb-2">
                                        {"Press Button To Run Code"}
                                    </Card.Subtitle>
                                </Card.Body>

                                <Card.Body
                                    style={{
                                        height: "60%",
                                        whiteSpace: "normal"
                                    }}>
                                    <Card.Text>
                                        The code of this submission has been made
                                        executable by the publisher(s). If allowed by the publisher(s), enter either your
                                        stdin, command line arguments or add a custom
                                        file to run with custom input. Please ensure your 
                                        input conforms with that required by the code.
                                    </Card.Text>
                                </Card.Body>
                                
                                <Card.Body>
                                {showButton && (
                                    <Button
                                    
                                    onClick={() => 
                                        { 
                                        isLoading
                                            ? runCode(submissionToken)
                                            : setShowMessage(true) 
                                        
                                        }}
                                    size="lg"
                                    
                                    >
                                    Run
                                    </Button>
                                )}
                                </Card.Body>
                               
                            </Tab>
                            
                            <Tab eventKey="userInput" title="Custom Input">
                                <Card.Body>
                                        <Card.Title>{"Enter Custom Input"}</Card.Title>
                                        <Card.Subtitle className="mb-2">
                                            {"Ensure Your Input Confomrs With That Required By The Code"}
                                        </Card.Subtitle>
                                    </Card.Body>

                                    <Card.Body
                                        style={{
                                            height: "60%",
                                            whiteSpace: "normal"
                                        }}>
                                        <Form>
                                        <Form.Group className="mb-3" controlId="stdin">
                                            <Form.Label>Standard Input</Form.Label>
                                            <Form.Control as="textarea" rows={3} />
                                        </Form.Group>
                                        <Form.Group className="mb-3" controlId="args">
                                            <Form.Label>Command Line Arguments</Form.Label>
                                            <Form.Control as="textarea" rows={3} />
                                        </Form.Group>
                                        <Form.Group controlId="formFileMultiple" className="mb-3">
                                            <Form.Label>Files</Form.Label>
                                            <Form.Control type="file" multiple />
                                        </Form.Group>
                                        </Form>
                                    </Card.Body>
                                    
                                    <Card.Body>
                                    {showButton && (
                                        <Button
                                        
                                        onClick={() => 
                                            { 
                                            isLoading
                                                ? runCode(submissionToken)
                                                : setShowMessage(true) 
                                            
                                            }}
                                        size="lg"
                                        
                                        >
                                        Run
                                        </Button>
                                    )}
                                    </Card.Body>
                            </Tab>
                            <Tab eventKey="advanced" title="Advanced">
                                 <Card.Body>
                                        <Card.Title>{"Advanced Settings"}</Card.Title>
                                        <Card.Subtitle className="mb-2">
                                            {""}
                                        </Card.Subtitle>
                                    </Card.Body>

                                    <Card.Body
                                        style={{
                                            height: "60%",
                                            whiteSpace: "normal"
                                        }}>
                                            <Form>
                                            <Form.Check 
                                                type="switch"
                                                id="memoryUsage"
                                                label="Get Memory Used by Program"
                                            />
                                            <Form.Check 
                                                type="switch"
                                                id="runTime"
                                                label="Get Run Time"
                                            />
                                            
                                        <Form.Group className="mb-3" controlId="numRuns">
                                            <Form.Label>Number of Runs</Form.Label>
                                            <Form.Control as="textarea" rows={1} />
                                        </Form.Group>
                                        <Card.Text>
                                        Run the program n number of runs and take average of run time and memory used by program.
                                    </Card.Text>
                                        <Form.Group className="mb-3" controlId="expectedOutput">
                                            <Form.Label>Expected Output</Form.Label>
                                            <Form.Control as="textarea" rows={1} />
                                        </Form.Group>
                                
                                        </Form>
                                    </Card.Body>
                                    
                                    <Card.Body>
                                    {showButton && (
                                        <Button
                                        
                                        onClick={() => 
                                            { 
                                            isLoading
                                                ? runCode(submissionToken)
                                                : setShowMessage(true) 
                                            
                                            }}
                                        size="lg"
                                        
                                        >
                                        Run
                                        </Button>
                                    )}
                                    </Card.Body>
                            </Tab>
                            <Tab eventKey="help" title="Help">
                            <Card.Body>
                                    <Card.Title>{"How To Run Submission"}</Card.Title>
                                </Card.Body>

                                <Card.Body
                                    style={{
                                        height: "60%",
                                        whiteSpace: "normal"
                                    }}>
                                    <Card.Text>
                                        Before attempting to run a submission, please ensure you understand the code and if any input is required.
                                        In order to run the submission code without any custom input press the run button. 
                                        If you have any problems running a submission that does not require input please contact the publisher(s).
                                        To run a submission with custom input please ensure you comply withe the input restrictions outlined by the code.
                                        In order to upload a custom input file, make sure to provide the run and if required compile bash scripts.
                                        For more infomration about utilizing a custom input file please see our "How To Make Your Submission Executable" section.
                                    </Card.Text>
                                </Card.Body>
                            </Tab>
                        </Tabs>
                        
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
                                {
                                    "Results"
                                }
                            </Alert.Heading>
                            <p>
                                Output: {output} 
                            </p>
                            <p>
                                Time: {runTime} sec
                            </p>
                            <p>
                                Memory: {memoryUsage} kB
                            </p>
                            <Button onClick={() =>  setShowMessage(false)}>
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
                :createSubmission(submissionToken)
                }
            </div>
        </div>
    )
}
