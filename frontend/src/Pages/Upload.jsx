/**
 * Upload.jsx
 * author: 190019931
 * 
 * Page for uploading files
 */

import React from "react";
import DragAndDrop from "./DragAndDrop";
import  axiosInstance from "../Web/axiosInstance";
import {Form, Button, Card, ListGroup, CloseButton, FloatingLabel} from "react-bootstrap";

const  uploadEndpoint = '/upload'

class Upload extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            files: [],
            submissionName: ""
        };

        this.dropFiles = this.dropFiles.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleDrop = this.handleDrop.bind(this);
        this.setSubmissionName = this.setSubmissionName.bind(this);
    }

    dropFiles(e) {
        this.handleDrop(e.target.files);
    }

    setSubmissionName(e) {
        this.setState({
            submissionName: e.target.value
        })
    }

    /**
     * Sends a POST request to the go server to upload (submission) files
     *
     * @param {JSON} userId Submission files' Author's User ID
     * @param {Array.<File>} files Submission files
     * @returns
     */
    uploadFiles(userId, submissionName, files) {

        if(userId === null) {
            console.log("not logged in!");
            return;
        }

        const authorID = JSON.parse(userId).userId;  //Extract author's userId

        const filePromises = files.map((file, i) => {   //Create Promise for each file (Encode to base 64 before upload)
            return new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.readAsDataURL(file);
                reader.onload = function(e) {
                    files[i] = e.target.result;
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
                       author: authorID,
                       name: submissionName,
                       content: files[i]
                   }
                   console.log(data)
                   axiosInstance.post(uploadEndpoint, data)
                                .then((response) => {
                                    console.log("Submission ID: " + response.data["projId"]);
                                    var codePage = window.open("/code");
                                    codePage.submission = submission;
                                    codePage.submissionName = this.state.submissionName;
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

        let userId = null;                          //Preparing to get userId from session cookie
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userId"){       //  If userId key exists, extract the userId value
                userId = cookie[1].trim();
                break;
            }
        }

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
        this.uploadFiles(userId, this.state.submissionName, this.state.files);

        document.getElementById("formFile").files = new DataTransfer().files;
        document.getElementById("submissionName").value = "";
        this.setState({
            files: [],
            submissionName: ""
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

        for(var i = 0; i < files.length; i++){
            if(!files[i] 
            || !(files[i].name.endsWith(".css") || files[i].name.endsWith(".java") || files[i].name.endsWith(".js"))){
                console.log("Invalid file");
                return;
            } 

            for(var j = 0; j < fileList.length; j++){
                if(files[i].name === fileList[j].name){
                    console.log("Duplicate file");
                    return;
                }
            }

            fileList.push(files[i]);
            formFileList.items.add(files[i]);
        }
        
        document.getElementById("formFile").files = formFileList.files;
        this.setState({
            files: fileList
        });
        
    }

    removeFile(key) {
        let formFileList = new DataTransfer();
        let fileList = this.state.files;

        for(var i = 0; i < this.state.files.length; i++){
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

		return (
            <div className="col-md-6 offset-md-5">
                <br/>

                <Form onSubmit={this.handleSubmit}>
                <DragAndDrop handleDrop={this.handleDrop}>
                    <Card style={{ width: '18rem' }}>
                    <Card.Header className="text-center"><h5>Upload Files</h5></Card.Header>
                        <Form.Group controlId="formFile" className="mb-3">
                            <Form.Control type="file" accept=".css,.java,.js" required onChange={this.dropFiles}/>{/* multiple later */}
                        </Form.Group>
                        <Card.Body>

                            {this.state.files.length > 0 ? (
                                <ListGroup>{files}</ListGroup>
                            ) : (
                                <Card.Text className="text-center" style={{color:"grey"}}><i>Drag and Drop <br/>here</i><br/><br/></Card.Text>
                            )}
                        </Card.Body>

                        <FloatingLabel controlId="submissionName" label="Submission name" className="mb-0">
                            <Form.Control type="text" placeholder="My_Submission" required onChange={this.setSubmissionName}/>
                        </FloatingLabel>
                        
                        <Card.Footer className="text-center"><Button variant="outline-secondary" type="submit">Upload files</Button>{' '}</Card.Footer>
                        
                    </Card>
                    </DragAndDrop>
                </Form>
                
            </div>
        )
	}
}

export default Upload;
