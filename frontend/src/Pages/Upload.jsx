/**
 * Upload.jsx
 * author: 190019931
 * 
 * Page for uploading files
 */

import React from "react";
import DragAndDrop from "./DragAndDrop";
import {Form, Button, Card, ListGroup, CloseButton, FloatingLabel} from "react-bootstrap"

class Upload extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            files: [],
            projectName: ""
        };

        this.dropFiles = this.dropFiles.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleDrop = this.handleDrop.bind(this);
        this.setProjectName = this.setProjectName.bind(this);
    }

    dropFiles(e) {
        this.handleDrop(e.target.files);
    }

    setProjectName(e) {
        this.setState({
            projectName: e.target.value
        })
    }

    handleSubmit(e) {
        e.preventDefault();

        //Checking there are files to submit
        if(this.state.files.length === 0){
            return;
        }

        let userID = null;                          //Preparing to get userID from session cookie
        let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
        for(let i = 0; i < cookies.length; i++){    //For each cookie,
            let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
            if(cookie[0].trim() == "userID"){       //  If userID key exists, extract the userID value
                userID = cookie[1].trim();
                break;
            }
        }

        if(userID === null){                        //If user has not logged in, disallow submit
            console.log("Not logged in");
            return;
        }

        // this.state.projectName = this.state.files[0].name;     //Temp, 1 file uploads
        // console.log(this.state.projectName);

        this.props.upload(userID, this.state.projectName, this.state.files).then((projectID) => {
            console.log("Project ID: " + projectID);
            var codePage = window.open("/code");
            codePage.projectID = projectID;
        }, (error) => {
            console.log(error);
        });
        

        document.getElementById("formFile").files = new DataTransfer().files;
        document.getElementById("projectName").value = "";
        this.setState({
            files: [],
            projectName: ""
        });

        console.log("Files submitted");
    }

    handleDrop(files) {
        // console.log(files);
        // console.log(this.state.files);
        // console.log(document.getElementById("formFile").files);

        if(this.state.files.length === 1) return; /* Remove later for multiple files */

        // if(this.writer.userID === null){
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

                        <FloatingLabel controlId="projectName" label="Project name" className="mb-0">
                            <Form.Control type="text" placeholder="My_Project" required onChange={this.setProjectName}/>
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