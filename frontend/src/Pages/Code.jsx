/**
 * Code.jsx
 * Author: 190019931
 * 
 * React component for displaying code
 */

import React, {useState, useEffect, useRef} from "react";
import axiosInstance from "../Web/axiosInstance";
import {Container, Row, Col, Form, InputGroup, Card, Button} from "react-bootstrap"
import MonacoEditor from 'react-monaco-editor';
import Comments from "./Comments";

const fileEndpoint = '/file'

function Code({id}) {

    const [file, setFile] = useState({ID:null, submissionId:null, path:"", CreatedAt:"", UpdatedAt:""});
    const [code, setCode] = useState("// type your code...");

    const monacoRef = useRef();
    const [theme, setTheme] = useState('vs');
    const [language, setLanguage] = useState('javascript');
    const [lineNumber, setLineNumber] = useState(1);

    const [comments, setComments] = useState([
		{CreatedAt: "", UpdatedAt: "", DeletedAt: "", base64Value: "TG9va3MgZ29vZCE=", line:1, ID:0, author: "John Doe", comments:[
			{CreatedAt: "", UpdatedAt: "", DeletedAt: "", base64Value: "SSBkaXNhZ3JlZS4=", line:1, ID:1, author: "Jane Doe", comments:[
				{CreatedAt: "", UpdatedAt: "", DeletedAt: "", base64Value: "SSBoYXZlIDUwMCBtb3JlIGNpdGF0aW9ucyB0aGFuIGJvdGggb2YgeW91LCBJIGNhbiBhc3N1cmUgeW91LCB0aGlzIGNvZGUgaXMgbWVkaW9jcmUu", line:1, ID:2, author: "Jim Doe", comments:[]}
			]}
		]},
		{CreatedAt: "", UpdatedAt: "", DeletedAt: "", base64Value: "VGhpcyBzZWVtcyBxdWl0ZSBpbmVmZmljaWVudC4=", line:2, ID:4, author: "Joe Doe", comments:[]},
	])
    const [showComments, setShowComments] = useState(false);

    useEffect(() => {
        getFile()
    }, [id])

    const getFile = () => {
        console.log(id);
        if(id == null || id == -1) return;
        axiosInstance.get(fileEndpoint + "/" + id)
            .then((response) => {
                console.log(response.data.file);

                //Set file, code and comments
                setFile(response.data.file);
                setCode(atob(response.data.file.base64Value));
                setComments(response.data.file.comments);

            }).catch((error) => {
                console.log(error);
            })
    }

    const editorDidMount = (editor, monaco) => {
        console.log('editorDidMount', editor);

        // monaco.languages.registerHoverProvider(language, {
        //     provideHover: function(model, position) {
        //         // console.log(model.getWordAtPosition(position).word); //Able to retrieve word
        //         // console.log(position); //Can also support any arbritrary range within code (Comment on lines/words)

        //         return {
        //             range: new monaco.Range(position.lineNumber, 1, model.getLineMaxColumn(position.lineNumber)),
        //             contents: [
        //                 { value: '**Comments**' },
        //                 { supportHtml: true, value: "[Reviewer Comments](http://localhost:23409/comment)"}
        //             ]
        //         }
        //     }
        // });

        editor.addAction({

            id: 'Comment',                                                  // An unique identifier of the contributed action.
            label: 'Comment',                                               // A label of the action that will be presented to the user. (Right-click)
            keybindings: [ monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyM ],   // An optional array of keybindings for the action.
            precondition: null,                                             // A precondition for this action.
            keybindingContext: null,                                        // A rule to evaluate on top of the precondition in order to dispatch the keybindings.
            contextMenuGroupId: 'navigation',
            contextMenuOrder: 1.5,
        
            // Method that will be executed when the action is triggered.
            // @param editor The editor instance is passed in as a convenience
            run: function (ed) {
                // let comment = prompt("Comment on line " + ed.getPosition().lineNumber, "Type Here");
                setLineNumber(ed.getPosition().lineNumber);
                setShowComments(true);
            }
        });
        // editor.deltaDecorations(
        //     [],
        //     [
        //         {
        //             range: new monaco.Range(1, 1, 1, 1),
        //             options: { 
        //                 isWholeLine: true,
        //                 className: 'myContentClass',
        //                 glyphMarginClassName: 'myGlyphMarginClass',
        //                 hoverMessage: [{value: "**Comments**"}, {value: "[link](#comments)"}],
        //                 glyphMarginHoverMessage: [{value: "Bye"}, {value: "[link](https://www.google.com)"}],
        //             }
        //         }
        //     ]
        // )
        editor.focus();
    };

    const onChange = (newValue, e) => {
        console.log('onChange', newValue, e);
        setCode(newValue)
    };

    const options = {
        selectOnLineNumbers: true,
        glyphMargin: true,
        readOnly: true
    };

    return(
        <Card border="light" className='row no-gutters'>
            <Card.Header><b>Code</b></Card.Header>
            <Card.Body>
            <Card.Title>{file.path}</Card.Title>
            <Card.Text>Created: {file.CreatedAt}</Card.Text>
            {file.path.split(".").pop() !== "pdf" ? 
                <Container fluid>
                        <Row>
                            <Col>
                                <InputGroup size="sm" className="mb-3">
                                    <InputGroup.Text id="inputGroup-sizing-sm">Language: </InputGroup.Text>
                                    <Form.Select defaultValue={language} size="sm" onChange={(e) => { setLanguage(e.target.value) }}>
                                        <option value="javascript">Javascript</option>
                                        <option value="html">HTML</option>
                                        <option value="css">CSS</option>
                                        <option value="json">JSON</option>
                                        <option value="java">Java</option>
                                        <option value="python">Python</option>
                                    </Form.Select>
                                </InputGroup>
                            </Col>
                            <Col>
                                <InputGroup size="sm" className="mb-3">
                                    <InputGroup.Text id="inputGroup-sizing-sm">Theme: </InputGroup.Text>
                                    <Form.Select size="sm" onChange={(e) => { setTheme(e.target.value) }}>
                                        <option value="vs">Visual Studio</option>
                                        <option value="vs-dark">Visual Studio Dark</option>
                                        <option value="hc-black">High Contrast Dark</option>
                                    </Form.Select>
                                </InputGroup>
                            </Col>
                        </Row>
                        <Row>
                            <Col>
                                <MonacoEditor
                                    ref={monacoRef}
                                    height="1000"
                                    language={language}
                                    theme={theme}
                                    value={code}
                                    options={options}
                                    onChange={onChange}
                                    editorDidMount={editorDidMount}
                                />
                            </Col>
                        </Row>
                </Container>
            :
                <embed height="1000" width="100%" src={"data:application/pdf;base64," + file.base64Value} />
            }
            <br/>
            <Button variant="dark" onClick={() => setShowComments(!showComments)}>Show comments</Button>
            <Comments 
                id={id}
                comments={comments}
                line={lineNumber}
                show={showComments}
                setShow={setShowComments}
                refresh={getFile}>
            </Comments>
            </Card.Body>
            <Card.Footer className="text-muted">Last updated: {file.UpdatedAt}</Card.Footer>
        </Card>
    )
}

export default Code;