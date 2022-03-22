/**
 * Code.jsx
 * Author: 190019931
 * 
 * React component for displaying code
 */

import React, {useState, useEffect, useRef} from "react";
import axiosInstance from "../Web/axiosInstance";
import {Container, Row, Col, Form, InputGroup, Card} from "react-bootstrap"
import MonacoEditor from 'react-monaco-editor';
import Comments from "./Comments";

const fileEndpoint = '/file'

function Code(props) {

    const [file, setFile] = useState({ID:null, submissionId:null, path:"", CreatedAt:"", UpdatedAt:""});
    const [code, setCode] = useState("// type your code...");

    const monacoRef = useRef();
    const [theme, setTheme] = useState('vs');
    const [language, setLanguage] = useState('javascript');
    const [lineNumber, setLineNumber] = useState(1);
    const [showComments, setShowComments] = useState(false);

    useEffect(() => {
        if(props.id == null) return;

        //Get File
        axiosInstance.get(fileEndpoint + "/" + props.id)
            .then((response) => {
                console.log(response.data);

                //Set file and code
                setFile(response.data);
                setCode(atob(response.data.base64Value));
            }).catch((error) => {
                console.log(error);
            })
    }, [props.id])

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
        //                 linesDecorationsClassName: 'myLineDecoration',
        //                 inlineClassName: 'myInlineDecoration',
        //                 hoverMessage: [{value: "Hello"}, {value: "[link](#comments)"}],
        //                 glyphMarginHoverMessage: [{value: "Bye"}, {value: "[link](https://www.google.com)"}],
        //                 glyphMarginClassName: 'myGlyphMarginClass'
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
        glyphMargin: true
        // readOnly: true
    };

    return(
        <Card border="light" className='row no-gutters'>
            <Card.Header><b>Code</b></Card.Header>
            <Card.Body>
            <Card.Title>{file.path}</Card.Title>
            <Card.Text>Created: {file.CreatedAt}</Card.Text>
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
                        <Comments id={props.id} line={lineNumber} show={showComments} setShow={setShowComments}></Comments>
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
            </Card.Body>
            <Card.Footer className="text-muted">Last updated: {file.UpdatedAt}</Card.Footer>
        </Card>
    )
}

export default Code;