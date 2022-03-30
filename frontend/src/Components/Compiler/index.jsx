import React, { useState, useEffect } from "react"
import {
	Card,
	Button,
	Badge,
	Container,
	Alert,
	Tab,
	Tabs,
	Form
} from "react-bootstrap"
import styles from "./Compiler.module.css"
import { useNavigate } from "react-router-dom"
import axios from "axios"
import { CSSTransition } from "react-transition-group"
import axiosInstance from "../../Web/axiosInstance"
import DragAndDrop from "../DragAndDrop/index"


const baseURL = "https://judge0-ce.p.rapidapi.com/submissions" //use wait param to only use one tag!!!!

const params = { base64_encoded: "true", wait: "true", fields: "*" }
const headers = {
	"content-type": "application/json",
	"Content-Type": "application/json",
	"X-RapidAPI-Host": "judge0-ce.p.rapidapi.com",
	"X-RapidAPI-Key": "488c3c10c9msh8121e54c23bb036p1afd03jsn6192663efef7"

}

const config = {
	headers: headers,
	params: params
}

export default ({ id }) => {
	const [zipAsBase64, setZipAsBase64] = useState(null)
	const [submissionToken, setSubmissionToken] = useState(null)
	const [output, setOutput] = useState(null)
	const [runTime, setRunTime] = useState(null)
	const [memoryUsage, setMemoryUsage] = useState(null)
	const [isLoading, setLoading] = useState(true)
	const [error, setError] = useState(null)
    const [submission, setSubmission] = useState({
        runnable: false,
        takesStdIn: false,
        takesCmdLn: false,
        takesInputFile: false,
        reqNetworkAccess: false,
    })

	useEffect(() => {
		getConfig(id)
		getZip(id)
	}, [id])

	//When no custom input file is allowed then we can just retrieve the zip as bas64 already
	const getZip = (id) => {
		axiosInstance
			.get("/submission/" + id + "/download")
			.then((response) => {
				//Set file, code and comments
				setZipAsBase64(response.data)
				console.log(response.data)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const getConfig = (id) => {
		axiosInstance
			.get("/submission/" + id)
			.then((response) => {
				//Set file, code and comments  
                setSubmission(response.data)
			})
			.catch((error) => {
				console.log(error)
			})
	}

	const createMultiFile = () => {
		return {
			language_id: 89,
			additional_files: zipAsBase64,
			stdInput: null,
			command_line_arguments: null
		}
	}
	const stdInputForm = () => {
		if (submission.takesStdIn)
			return (
				<div>
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
				</div>
			)
	}

    const argsForm = () => {
        if(submission.takesCmdLn)
        return (
            <div> 
                <Form.Group className="mb-3" controlId="args">
							<Form.Label>Command Line Arguments</Form.Label>
							<Form.Control as="textarea" rows={3} />
						</Form.Group>
            </div>
        )
    }

    const inputFileForm = () => {
        if(submission.takesInputFile)
        return(
            <div>
                <Form.Group
							controlId="formFileMultiple"
							className="mb-3">
							<Form.Label>Files</Form.Label>
							<Form.Control type="file" multiple />
						</Form.Group>
            </div>
        )
    }

	const customInput = (showButton) => {
		return (
			<Tab eventKey="userInput" title="Custom Input">
				<Card.Body>
					<Card.Title>{"Enter Custom Input"}</Card.Title>
					<Card.Subtitle className="mb-2">
						{
							"Ensure Your Input Confomrs With That Required By The Code"
						}
					</Card.Subtitle>
				</Card.Body>
				<Card.Body
					style={{
						height: "60%",
						whiteSpace: "normal"
					}}>
					<Form>
						{stdInputForm()}
						{argsForm()}
                        {inputFileForm()}
					</Form>
				</Card.Body>

				<Card.Body>
					{showButton && (
						<Button
							onClick={() => {
								isLoading
									? runCode(multiFileData)
									: setShowMessage(true)
							}}
							size="lg">
							Run
						</Button>
					)}
				</Card.Body>
			</Tab>
		)
	}

	const runCode = (multiFileData) => {
		getZip(id)
		let data = createMultiFile(zipAsBase64)
		axios
			.post(baseURL, data, config)
			.then(function (response) {

				setCompilerResponse(response)
				setSubmissionToken(response.data.token)
				setOutput(window.atob(response.data.stdout))
				setMemoryUsage(response.data.memory)
				setRunTime(response.data.time)
				setLoading(false)
			})
			.catch(function (error) {
				console.log(error)
				setLoading(false)
			})
	}

	const createSubmission = (submission) => {
		const [showButton, setShowButton] = useState(true)
		const [showMessage, setShowMessage] = useState(false)
		const [key, setKey] = useState("run")
		return (
			<div>
				<Card style={{ marginTop: "8px" }} className="rounded">
					<Tabs
						id="controlled-tab-example"
						activeKey={key}
						onSelect={(k) => setKey(k)}
						className="mb-3">
						<Tab eventKey="run" title="Run">
							<Card.Body>
								<Card.Title>
									{"This Code Can Be Run"}
								</Card.Title>
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
									executable by the publisher(s). If allowed
									by the publisher(s), enter either your
									stdin, command line arguments or add a
									custom file to run with custom input. Please
									ensure your input conforms with that
									required by the code.
								</Card.Text>
							</Card.Body>

							<Card.Body>
								{showButton && (
									<Button
										onClick={() => {
											isLoading
												? runCode(multiFileData)
												: setShowMessage(true)
										}}
										size="lg">
										Run
									</Button>
								)}
							</Card.Body>
						</Tab>
						{takesStdIn || takesCmdLn || takesInputFile
							? customInput(showButton)
							: console.log("No Custom Input")}
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

									<Form.Group
										className="mb-3"
										controlId="numRuns">
										<Form.Label>Number of Runs</Form.Label>
										<Form.Control as="textarea" rows={1} />
									</Form.Group>
									<Card.Text>
										Run the program n number of runs and
										take average of run time and memory used
										by program.
									</Card.Text>
									<Form.Group
										className="mb-3"
										controlId="expectedOutput">
										<Form.Label>Expected Output</Form.Label>
										<Form.Control as="textarea" rows={1} />
									</Form.Group>
								</Form>
							</Card.Body>

							<Card.Body>
								{showButton && (
									<Button
										onClick={() => {
											isLoading
												? runCode(multiFileData)
												: setShowMessage(true)
										}}
										size="lg">
										Run
									</Button>
								)}
							</Card.Body>
						</Tab>
						<Tab eventKey="help" title="Help">
							<Card.Body>
								<Card.Title>
									{"How To Run Submission"}
								</Card.Title>
							</Card.Body>

							<Card.Body
								style={{
									height: "60%",
									whiteSpace: "normal"
								}}>
								<Card.Text>
									Before attempting to run a submission,
									please ensure you understand the code and if
									any input is required. In order to run the
									submission code without any custom input
									press the run button. If you have any
									problems running a submission that does not
									require input please contact the
									publisher(s). To run a submission with
									custom input please ensure you comply withe
									the input restrictions outlined by the code.
									In order to upload a custom input file, make
									sure to provide the run and if required
									compile bash scripts. For more infomration
									about utilizing a custom input file please
									see our "How To Make Your Submission
									Executable" section.
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
					onExited={() => setShowButton(true)}>
					<Alert
						variant="primary"
						dismissible
						onClose={() => setShowMessage(false)}>
						<Alert.Heading>{"Results"}</Alert.Heading>
						<p>Output: {output}</p>
						<p>Time: {runTime} sec</p>
						<p>Memory: {memoryUsage} kB</p>
						<Button onClick={() => setShowMessage(false)}>
							Close
						</Button>
					</Alert>
				</CSSTransition>
			</div>
		)
	}

	return (
		<div>
			<div>
				{error
					? "Something went wrong, please try again later..."
					: submission.runnable
					? createSubmission(submissionToken)
					: console.log("Not Runnable")}
			</div>
		</div>
	)
}
