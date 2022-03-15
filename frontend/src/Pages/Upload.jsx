/**
 * Upload.jsx
 * author: 190019931
 * 
 * Page for uploading files
 */

import React from "react";
import DragAndDrop from "./DragAndDrop";
import  axiosInstance from "../Web/axiosInstance";
import {Form, Button, Card, ListGroup, CloseButton, FloatingLabel, Container, Row, Col, InputGroup, FormControl, Tabs, Tab} from "react-bootstrap";
import JwtService from "../Web/jwt.service"

const  uploadEndpoint = '/submissions/create'

class Upload extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            authors: [], //Change to string
            files: [],
            submissionName: "",
            submissionAbstract: "",
            tags: []
        };

        this.dropFiles = this.dropFiles.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleDrop = this.handleDrop.bind(this);
        this.setSubmissionName = this.setSubmissionName.bind(this);
        this.setSubmissionAbstract = this.setSubmissionAbstract.bind(this);
        this.tagsInput = React.createRef();
    }

    componentDidMount() {
        console.log(JwtService.getUserID());
    }

    dropFiles(e) {
        this.handleDrop(e.target.files);
    }

    setSubmissionName(e) {
        this.setState({
            submissionName: e.target.value
        })
    }

    setSubmissionAbstract(e) {
        this.setState({
            submissionAbstract: e.target.value
        })
    }

    /**
     * Sends a POST request to the go server to upload a submission
     *
     * @param {JSON} authors Submission files' Authors' User ID
     * @param {Array.<File>} files Submission files
     * @returns
     */
    uploadSubmission(authors, submissionName, submissionAbstract, files, categories) {

        console.log(authors);
        // const authorID = JSON.parse(userId).userId;  //Extract author's userId

        let files2 = [];
        const filePromises = files.map((file, i) => {   //Create Promise for each file (Encode to base 64 before upload)
            return new Promise((resolve, reject) => {
                files2.push({
                    name: files[i].name,
                    path: files[i].name,
                });
                files[i].path = files[i].name;
                const reader = new FileReader();
                reader.readAsDataURL(file);
                reader.onload = function(e) {
                    files[i].base64Value = e.target.result.split(',')[1];
                    files2[i].base64Value = e.target.result.split(',')[1];
                    resolve();                          //Promise(s) resolved/fulfilled once reader has encoded file(s) into base 64
                }
                reader.onerror = function() {
                    reject();
                }
            });
        })

        // return new Promise ( (resolve, reject) => {
        //    var request = new XMLHttpRequest();
        //    Promise.all(filePromises) //Encode all files to upload into base 64 before uploading
        //        .then(() => {
        //            let data = {
        //                author : authorID,
        //                name : submissionName,
        //                content : files[0]
        //            }
        //            console.log(data);
                    // get a callback when the server responds
        //            request.addEventListener('load', () => {
                        //Return response for code page
        //                resolve(request.responseText);
        //            })
        //            request.onerror = reject;
                    // open the request with the verb and the url TEMP: this will potentially change with actual URL
        //            request.open('POST', BACKEND_ADDRESS + uploadEndpoint)
        //            this.sendSecureRequest(request, data)
        //        })
        //        .catch((error) => {
        //            console.log(error);
        //        })
        // })
        Promise.all(filePromises)
               .then(() => {
                   let data = {
                        name: submissionName,
                        license: "MIT",
                        abstract: submissionAbstract,
                        files: files2,
                        authors: authors,
                        categories: categories
                   }
                   console.log(data)
                   console.log(data.files[0].name)
                   axiosInstance.post(uploadEndpoint, data)
                                .then((response) => {
                                    console.log(response);
                                    // console.log("Submission ID: " + response.data["id"]);
                                    // var codePage = window.open("/code");
                                    // codePage.submissionId = response.data["id"];
                                    // codePage.submissionName = response.data["name"];
                                    // codePage.submission = files[0];
                                })
                                .catch((error) => {
                                    console.log(error);
                                })
               })
    }

    handleSubmit(e) {
        e.preventDefault();

        //Checking there are files to submit
        if(this.state.files.length === 0){
            return;
        }

        let userId = JwtService.getUserID();        //Preparing to get userId

        if(userId === null){                        //If user has not logged in, disallow submit
            console.log("Not logged in");
            return;
        }

        // this.state.submissionName = this.state.files[0].name;     //Temp, 1 file uploads
        // console.log(this.state.submissionName);

        // this.props.upload(userId, this.state.submissionName, this.state.files).then((submission) => {
        //    console.log("Submission ID: " + submission);
        //    var codePage = window.open("/code");
        //    codePage.submission = submission;
        //    codePage.submissionName = this.state.submissionName;
        // }, (error) => {
        //    console.log(error);
        // });
        
        this.state.authors.push(userId);
        this.uploadSubmission(this.state.authors, this.state.submissionName, this.state.submissionAbstract, this.state.files, this.state.tags);

        //Clearing form
        document.getElementById("formFile").files = new DataTransfer().files;
        document.getElementById("submissionName").value = "";
        this.setState({
            authors: [],
            files: [],
            submissionName: "",
            submissionAbstract: "",
            tags: []
        });

        console.log("Files submitted");
    }

    handleDrop(files) {
        // console.log(files);
        // console.log(this.state.files);
        // console.log(document.getElementById("formFile").files);

        if(this.state.files.length === 1) return; /* Remove later for multiple files */

        // if(this.writer.userId === null){
        //     console.log("Not logged in!");
        //     return;
        // }

        let formFileList = new DataTransfer();
        let fileList = this.state.files;

        for(let i = 0; i < files.length; i++){
            if(!files[i] 
            // || !(files[i].name.endsWith(".css") || files[i].name.endsWith(".java") || files[i].name.endsWith(".js"))
            ){
                console.log("Invalid file");
                return;
            } 

            for(let j = 0; j < fileList.length; j++){
                if(files[i].name === fileList[j].name){
                    console.log("Duplicate file");
                    return;
                }
            }

            console.log(files[i]);
            fileList.push(files[i]);
            formFileList.items.add(files[i]);
        }
        
        document.getElementById("formFile").files = formFileList.files;
        this.setState({ files: fileList });
        
    }

    removeFile(key) {
        let formFileList = new DataTransfer();
        let fileList = this.state.files;

        for(let i = 0; i < this.state.files.length; i++){
            formFileList.items.add(this.state.files[i]);
        }
        formFileList.items.remove(key);
        fileList.splice(key, 1);

        document.getElementById("formFile").files = formFileList.files;
        this.setState({
            files: fileList
        });
    }

	render() {

        const files = this.state.files.map((file, i) => {
            return (
                <ListGroup.Item as="li" key={i} action onClick={() => {}}>
                    <CloseButton onClick={() => {this.removeFile(i)}}/>
                    <br/>
                    <label>File name: {file.name}</label>
                    <br/>
                    <label>File type: {file.type}</label>
                    <br/>
                    <label>File Size: {file.size} bytes</label>
                    <br/>
                    <label>Last modified: {new Date(file.lastModified).toUTCString()}</label>
                </ListGroup.Item>
            );
        });

        const tags = this.state.tags.map((tag, i) => {
            return (
                <Button key={i} variant="outline-secondary" size="sm" onClick={() => {
                    this.setState({
                        tags: this.state.tags.filter(value => value !== tag)
                    })}
                }>
                    {tag}
                </Button>
            );
        });

		return (
            <Container>
                <br/>
                <Row>
                    <Col></Col>
                    <Col xs={4}>
                        <DragAndDrop handleDrop={this.handleDrop}>
                            <Card>
                                <Form onSubmit={this.handleSubmit}>
                                    <Card.Header className="text-center"><h5>Submission Upload</h5></Card.Header>
                                    <Tabs justify defaultActiveKey="details" id="profileTabs" className="mb-3">
                                        <Tab eventKey="details" title="Details">
                                            <Row>
                                            <Form.Group className="mb-3" controlId="submissionName">
                                                <Form.Label>Name</Form.Label>
                                                <Form.Control type="text" placeholder="Name" required onChange={this.setSubmissionName}/>
                                            </Form.Group>
                                            <Form.Group className="mb-3" controlId="submissionAbstract">
                                                <Form.Label>Abstract</Form.Label>
                                                <Form.Control as="textarea" rows={3} placeholder="Abstract" required onChange={this.setSubmissionAbstract}/>
                                            </Form.Group>
                                            </Row>
                                            <Row>
                                            <InputGroup className="mb-3">
                                                <FormControl
                                                    placeholder="Add tags here"
                                                    aria-label="Tags"
                                                    aria-describedby="addTag"
                                                    ref = {this.tagsInput}
                                                />
                                                <Button variant="outline-secondary" id="addTag" onClick={ () => {if(this.state.tags.includes(this.tagsInput.current.value)) return; this.setState({tags:[ ...this.state.tags, this.tagsInput.current.value]}); this.tagsInput.current.value = ""} }>
                                                Add
                                                </Button>
                                            </InputGroup>
                                            <Col>
                                                {tags}
                                            </Col>
                                        </Row>
				                        </Tab>
                                    <Tab eventKey="files" title="Files">
                                        <Row>
                                            <Form.Group controlId="formFile" className="mb-3">
                                                <Form.Control type="file" accept=".css,.java,.js" required onChange={this.dropFiles}/>{/* multiple later w/ "zip,application/octet-stream,application/zip,application/x-zip,application/x-zip-compressed" */}
                                            </Form.Group>
                                            <Card.Body>

                                                {this.state.files.length > 0 ? (
                                                    <ListGroup>{files}</ListGroup>
                                                ) : (
                                                    <Card.Text className="text-center" style={{color:"grey"}}><i>Drag and Drop <br/>here</i><br/><br/></Card.Text>
                                                )}
                                            </Card.Body>
                                        </Row>
                                    </Tab>
                                    </Tabs>
                                    <Card.Footer className="text-center"><Button variant="outline-secondary" type="submit">Upload</Button>{' '}</Card.Footer>
                                </Form>
                                </Card>
                        </DragAndDrop>
                    </Col>
                    <Col></Col>
                </Row>
            </Container>
        )
	}
}

export default Upload;
