import React from "react";
import DragAndDrop from "./DragAndDrop";

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

        this.setState({
            files: []
        });

        console.log("Files submitted");
    }

    handleDrop(files) {
        console.log(files);
        console.log( this.state.files);

        let formFileList = new DataTransfer();
        let fileList = this.state.files;

        for(var i = 0; i < files.length; i++){
            if(!files[i] || !files[i].name.endsWith(".zip")){
                console.log("Invalid file");
                return;
            } 
            fileList.push(files[i]);
            formFileList.items.add(files[i]);
        }
        
        document.getElementById("uploadFiles").files = formFileList.files;
        this.setState({
            files: fileList
        });
        
    }

	render() {

        const files = this.state.files.map((file, i) => {
            return (
                <button type="button" class="list-group-item list-group-item-action" disabled key={i}>
                    <label>File name: {file.name}</label>
                    <br/>
                    <label>File type: {file.type}</label>
                    <br/>
                    <label>File Size: {file.size} bytes</label>
                    <br/>
                    <label>Last modified: {new Date(file.lastModified).toUTCString()}</label>
                </button>
            );
        });

		return (
            <div className="col-md-6 col-md-offset-3">
                <head>
					<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous"/>
				</head>
                <h2>Upload File</h2>
                
                <form name="form" onSubmit={this.handleSubmit}>
                <DragAndDrop handleDrop={this.handleDrop}>
                    <div class="custom-file">
                        <label class="custom-file-label" htmlFor="uploadFiles">Choose/Drop files (.zip)</label>
                        <input type="file" class="custom-file-input" id="uploadFiles" name="uploadFiles" accept=".zip" onChange={this.handleChange} multiple/>
                    </div>

                    {/* <label htmlFor="uploadFiles" style={lblCSS}>Choose file(s) to upload (.zip)</label>
                    <input type="file" id="uploadFiles" name="uploadFiles" accept=".zip" onChange={this.handleChange} style={{opacity:0}} multiple/> */}

                    <div class="card">
                        <div class="card-body">
                            {this.state.files.length > 0 ? (
                                <ul class="list-group">{files}</ul>
                            ) : (
                                <label>No files selected.</label>
                            )}
                        </div>
                    </div>
                    </DragAndDrop>
                    <div>
                        <button className="btn btn-primary">Upload files</button>
                    </div>
                </form>
                
            </div>
        )
	}
}

export default Upload;