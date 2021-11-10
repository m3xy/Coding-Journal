/**
 * Upload.jsx
 * author: 190019931
 * 
 * Page for uploading files
 */

import React from "react";
import DragAndDrop from "./DragAndDrop";
import {Form, Button, Card, ListGroup, CloseButton} from "react-bootstrap"

class Upload extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            files: []
        };

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleDrop = this.handleDrop.bind(this);
    }

    handleChange(e) {
        this.handleDrop(e.target.files);
    }

    handleSubmit(e) {
        e.preventDefault();

        //Checking there are files to submit
        if(this.state.files.length === 0){
            return;
        }

        //Printing contents of each file
        this.state.files.forEach(
            file => {
                if (file) {
                    var reader = new FileReader();
                    reader.readAsText(file, "UTF-8");
                    reader.onload = function (e) {
                        console.log(e.target.result);
                    }
                    reader.onerror = function (e) {
                        console.log("Error reading file");
                    }
                }
            }
        );

        document.getElementById("formFile").files = new DataTransfer().files;
        this.setState({
            files: []
        });

        console.log("Files submitted");
    }

    handleDrop(files) {
        // console.log(files);
        // console.log(this.state.files);
        // console.log(document.getElementById("formFile").files);

        let formFileList = new DataTransfer();
        let fileList = this.state.files;

        for(var i = 0; i < files.length; i++){
            if(!files[i] || !files[i].name.endsWith(".zip")){
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

                <form name="form" onSubmit={this.handleSubmit}>
                <DragAndDrop handleDrop={this.handleDrop}>
                    <Card style={{ width: '18rem' }}>
                    <Card.Header className="text-center"><h5>Upload Files</h5></Card.Header>
                        <Form.Group controlId="formFile" className="mb-3">
                            <Form.Control type="file" accept=".zip" onChange={this.handleChange} multiple/>
                        </Form.Group>
                        <Card.Body>

                            {this.state.files.length > 0 ? (
                                <ListGroup>{files}</ListGroup>
                            ) : (
                                <Card.Text className="text-center" style={{color:"grey"}}><i>Drag and Drop <br/>here</i><br/><br/></Card.Text>
                            )}
                        </Card.Body>
                        <Card.Footer className="text-center"><Button variant="outline-secondary" type="submit">Upload files</Button>{' '}</Card.Footer>
                        
                    </Card>
                    </DragAndDrop>
                </form>
                
            </div>
        )
	}
}

export default Upload;